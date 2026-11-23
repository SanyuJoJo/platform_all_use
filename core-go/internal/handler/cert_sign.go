package handler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yourorg/core-go/internal/audit"
	"github.com/yourorg/core-go/internal/config"
	"github.com/yourorg/core-go/internal/envelope"
	"github.com/yourorg/core-go/internal/openssl"
	"github.com/yourorg/core-go/internal/pathguard"
)

// CertSignHandler cert.sign 终端证书签发。
//
// ★ 路径约定：
//   - output_dir 非空：
//       <output_dir>/<pubkey_sm3>.cert.pem
//       <output_dir>/<pubkey_sm3>.key.pem     白盒密文（若 core 持有私钥）
//   - output_dir 为空：data/certs/<cert_id>.pem（旧行为）
//
// ★ 不再生成 -chain.pem：
//   证书校验依赖 <CertRoot>/ca/<domain>/ 下的 rehash 文件（<hash>.0）。
type CertSignHandler struct{}

func (h *CertSignHandler) Execute(
	ctx context.Context, req *envelope.Request, cfg *config.Config,
) *envelope.Response {
	start := time.Now()

	// ---------- 白名单 ----------
	wl, err := config.GetWhitelist(cfg.WhitelistPath)
	if err != nil {
		return genErr("INTERNAL_ERROR", "whitelist load failed")
	}
	alg := strParamDefault(req.Params, "algorithm", "SM2")
	if !wl.IsAlgorithmAllowed("cert.sign", alg) {
		audit.Log(req, "cert.sign", start, "FAILED", "ALGORITHM_NOT_ALLOWED")
		return genErr("ALGORITHM_NOT_ALLOWED", "algorithm not allowed")
	}

	caSource := strParamDefault(req.Params, "ca_source", "local")
	if !wl.IsCAKindAllowed("cert.sign", caSource) {
		return genErr("INVALID_PARAM", "invalid ca_source: "+caSource)
	}

	// 证书类型支持多选（前端可同时申请 signature + encryption）。
	// ★ 必须按「全部选中类型」求 keyUsage 并集：国密 NTLS 允许同一张证书
	//   同时用于签名与加密，但证书必须同时具备 digitalSignature 与
	//   keyEncipherment 用途，否则 tongsuo 的 SSL_CTX_use_enc_certificate()
	//   会以 "called a function you should not call" 拒绝加载。
	certTypes := certTypesFromParams(req.Params)
	if len(certTypes) == 0 {
		certTypes = []string{"server"}
	}
	for _, ct := range certTypes {
		if !wl.IsCertTypeAllowed("cert.sign", ct) {
			return genErr("INVALID_PARAM", "invalid cert_type: "+ct)
		}
	}

	// ---------- 输出路径参数 ----------
	outputDir := strParam(req.Params, "output_dir")
	outputLayout := strParamDefault(req.Params, "output_layout", "server_pubkey_sm3")

	// ---------- 临时目录 ----------
	tmpDir, err := mkTmpDir("cert-sign")
	if err != nil {
		return genErr("CORE_EXEC_FAILED", err.Error())
	}
	defer cleanupTmpDir(tmpDir)

	// ---------- 解析 CA 来源 ----------
	var caCertAbs, caKeyAbs string
	switch caSource {
	case "local":
		caCertRel := strParam(req.Params, "ca_cert_path")
		caKeyRef := strParam(req.Params, "ca_key_ref")
		if caCertRel == "" || caKeyRef == "" {
			return genErr("INVALID_PARAM", "missing ca_cert_path or ca_key_ref")
		}
		caCertAbs, err = pathguard.Resolve(caCertRel)
		if err != nil {
			return genErr("PATH_NOT_ALLOWED", err.Error())
		}
		if _, err := os.Stat(caCertAbs); err != nil {
			return genErr("CERT_NOT_FOUND", "ca cert not found")
		}
		caKeyAbs, err = loadDecryptedKey(caKeyRef, tmpDir)
		if err != nil {
			return genErr("KEY_NOT_FOUND", err.Error())
		}
	case "manual":
		caCertRel := strParam(req.Params, "ca_cert_path")
		caKeyRel := strParam(req.Params, "ca_key_path")
		if caCertRel == "" || caKeyRel == "" {
			return genErr("INVALID_PARAM", "missing manual ca cert/key")
		}
		caCertAbs, err = pathguard.Resolve(caCertRel)
		if err != nil {
			return genErr("PATH_NOT_ALLOWED", err.Error())
		}
		caKeyAbs, err = pathguard.Resolve(caKeyRel)
		if err != nil {
			return genErr("PATH_NOT_ALLOWED", err.Error())
		}
	default:
		return genErr("INVALID_PARAM", "invalid ca_source")
	}

	client, err := openssl.NewClient()
	if err != nil {
		return genErr("CORE_EXEC_FAILED", err.Error())
	}

	// ---------- 解析 CSR 来源 ----------
	var (
		csrPath    string
		keyPath    string
		hasPrivKey bool
	)

	csrSrc := strParam(req.Params, "csr_source")
	switch csrSrc {
	case "existing":
		csrRel := strParam(req.Params, "csr_path")
		csrKeyRef := strParam(req.Params, "csr_key_ref")
		if csrRel == "" {
			return genErr("INVALID_PARAM", "missing csr_path for existing source")
		}
		csrPath, err = pathguard.Resolve(csrRel)
		if err != nil {
			return genErr("PATH_NOT_ALLOWED", "csr_path: "+err.Error())
		}
		if _, err := os.Stat(csrPath); err != nil {
			return genErr("CERT_NOT_FOUND", "csr not found")
		}
		if csrKeyRef != "" {
			keyPath, err = loadDecryptedKey(csrKeyRef, tmpDir)
			if err != nil {
				return genErr("KEY_NOT_FOUND", err.Error())
			}
			hasPrivKey = true
		}
	case "upload":
		csrRel := strParam(req.Params, "csr_path")
		if csrRel == "" {
			return genErr("INVALID_PARAM", "missing csr_path for upload source")
		}
		csrPath, err = pathguard.Resolve(csrRel)
		if err != nil {
			return genErr("PATH_NOT_ALLOWED", "csr_path: "+err.Error())
		}
		if _, err := os.Stat(csrPath); err != nil {
			return genErr("CERT_NOT_FOUND", "csr not found")
		}
		keyRel := strParam(req.Params, "csr_key_path")
		if keyRel != "" {
			keyPath, err = pathguard.Resolve(keyRel)
			if err != nil {
				return genErr("PATH_NOT_ALLOWED", "csr_key_path: "+err.Error())
			}
			hasPrivKey = true
		}
	case "":
		keyPath = filepath.Join(tmpDir, "leaf.key")
		if err := genPrivateKey(ctx, alg, req.Params, keyPath); err != nil {
			return genErr("CORE_EXEC_FAILED", "gen key: "+err.Error())
		}
		hasPrivKey = true

		subject, _ := subjectFromParams(req.Params)
		subjArg := subjectToOpenSSL(subject)
		csrPath = filepath.Join(tmpDir, "leaf.csr")
		if _, err := client.Run(ctx, "req", "-new", "-utf8", "-key", keyPath,
			"-out", csrPath, "-subj", subjArg); err != nil {
			return genErr("CORE_EXEC_FAILED", "csr: "+err.Error())
		}
	default:
		return genErr("INVALID_PARAM", "csr_source must be empty, existing or upload")
	}

	// ---------- 扩展配置 ----------
	extPath := filepath.Join(tmpDir, "ext.cnf")
	if err := writeCertExt(extPath, certTypes); err != nil {
		return genErr("CORE_EXEC_FAILED", err.Error())
	}

	// ---------- 输出路径决策 ----------
	certID := newID("cert")
	var (
		certAbs      string
		certRespPath string
		useNewLayout bool
	)

	if outputDir != "" {
		useNewLayout = true
		if !filepath.IsAbs(outputDir) {
			return genErr("INVALID_PARAM", "output_dir must be an absolute path")
		}
		if err := os.MkdirAll(outputDir, 0750); err != nil {
			return genErr("CORE_EXEC_FAILED", "mkdir output_dir: "+err.Error())
		}
		certAbs = filepath.Join(outputDir, certID+".cert.pem.tmp")
	} else {
		certRel := "data/certs/" + certID + ".pem"
		var rerr error
		certAbs, rerr = pathguard.Resolve(certRel)
		if rerr != nil {
			return genErr("PATH_NOT_ALLOWED", rerr.Error())
		}
		if err := pathguard.EnsureDir(filepath.Dir(certAbs), 0750); err != nil {
			return genErr("CORE_EXEC_FAILED", err.Error())
		}
		certRespPath = certRel
	}

	// ---------- 签发 ----------
	digest := digestForCertAlg(ctx, client, caCertAbs)
	days := intParam(req.Params, "validity_days", 365)
	if _, err := client.Run(ctx, "x509", "-req", "-in", csrPath,
		"-CA", caCertAbs, "-CAkey", caKeyAbs, "-CAcreateserial",
		"-out", certAbs, "-days", fmt.Sprintf("%d", days),
		digest, "-extfile", extPath); err != nil {
		return genErr("CORE_EXEC_FAILED", "sign: "+err.Error())
	}
	_ = os.Chmod(certAbs, 0640)

	// ---------- 新规范：重命名 + 白盒私钥（★ 不再生成 chain） ----------
	var (
		pubKeySM3   string
		keyRespPath string
	)

	if useNewLayout {
		// 1. 算 pubkey_sm3
		pubKeySM3, err = pubkeySM3FromCertAbs(ctx, client, certAbs, tmpDir)
		if err != nil {
			_ = os.Remove(certAbs)
			return genErr("CORE_EXEC_FAILED", "calc pubkey_sm3: "+err.Error())
		}

		// 2. 重命名证书 → <pubkey_sm3>.cert.pem
		finalCertAbs := filepath.Join(outputDir, pubKeySM3+".cert.pem")
		if err := os.Rename(certAbs, finalCertAbs); err != nil {
			_ = os.Remove(certAbs)
			return genErr("CORE_EXEC_FAILED", "rename cert: "+err.Error())
		}
		certAbs = finalCertAbs
		certRespPath = certAbs

		// 3. 白盒落盘终端私钥（★ 三件套只剩两件：cert + key）
		if hasPrivKey && keyPath != "" {
			finalKeyAbs := filepath.Join(outputDir, pubKeySM3+".key.pem")
			if err := whiteboxEncrypt(ctx, cfg, keyPath, finalKeyAbs); err != nil {
				_ = os.Remove(certAbs)
				return genErr("CORE_EXEC_FAILED", "whitebox encrypt key: "+err.Error())
			}
			keyRespPath = finalKeyAbs
		}
	}

	// ---------- 旧规范：私钥走 keystore ----------
	var keyRef string
	if !useNewLayout && hasPrivKey && keyPath != "" {
		plainKey, _ := os.ReadFile(keyPath)
		keyRef, err = saveEncryptedKey(plainKey)
		if err != nil {
			return genErr("CORE_EXEC_FAILED", "save leaf key: "+err.Error())
		}
	}

	// ---------- 审计 + 响应（★ 无 chain_path） ----------
	audit.Log(req, "cert.sign", start, "SUCCESS", "")

	data := map[string]interface{}{
		"cert_id":   certID,
		"cert_path": certRespPath,
		"serial":    extractSerial(ctx, client, certAbs),
	}
	if useNewLayout {
		data["pubkey_sm3"] = pubKeySM3
		data["output_layout"] = outputLayout
		if hasPrivKey {
			data["key_path"] = keyRespPath
		}
	} else {
		if keyRef != "" {
			data["key_ref"] = keyRef
		}
	}

	return envelope.NewSuccess(data)
}

// certTypesFromParams 解析证书类型列表。
//
// 优先读取 cert_types（数组或逗号分隔字符串，前端多选场景），
// 缺省回退到 cert_type（单值，兼容旧调用）。
func certTypesFromParams(params map[string]interface{}) []string {
	var out []string
	seen := make(map[string]bool, 4)
	add := func(raw string) {
		t := strings.ToLower(strings.TrimSpace(raw))
		if t == "" || seen[t] {
			return
		}
		seen[t] = true
		out = append(out, t)
	}
	if v, ok := params["cert_types"]; ok {
		switch x := v.(type) {
		case []interface{}:
			for _, item := range x {
				if s, ok := item.(string); ok {
					add(s)
				}
			}
		case []string:
			for _, s := range x {
				add(s)
			}
		case string:
			for _, s := range strings.Split(x, ",") {
				add(s)
			}
		}
	}
	if len(out) == 0 {
		add(strParam(params, "cert_type"))
	}
	return out
}

// writeCertExt 写证书扩展配置文件。
//
// certTypes 为全部申请类型，keyUsage 取并集：
//
//	signature  → digitalSignature,nonRepudiation
//	encryption → keyEncipherment,keyAgreement
//	server     → digitalSignature,keyEncipherment（+ EKU serverAuth）
//	client     → digitalSignature（+ EKU clientAuth）
//
// 同时申请 signature + encryption 时两类用途叠加：这样同一张证书既能用于
// 国密签名也能用于国密加密（tongsuo 要求加密证书带 keyEncipherment /
// dataEncipherment，否则 SSL_CTX_use_enc_certificate 会拒绝加载）。
func writeCertExt(path string, certTypes []string) error {
	var ku, eku []string
	seenKU := make(map[string]bool, 4)
	seenEKU := make(map[string]bool, 2)
	addKU := func(items ...string) {
		for _, it := range items {
			if it = strings.TrimSpace(it); it != "" && !seenKU[it] {
				seenKU[it] = true
				ku = append(ku, it)
			}
		}
	}
	addEKU := func(items ...string) {
		for _, it := range items {
			if it = strings.TrimSpace(it); it != "" && !seenEKU[it] {
				seenEKU[it] = true
				eku = append(eku, it)
			}
		}
	}

	for _, t := range certTypes {
		switch t {
		case "server":
			addKU("digitalSignature", "keyEncipherment")
			addEKU("serverAuth")
		case "client":
			addKU("digitalSignature")
			addEKU("clientAuth")
		case "signature":
			addKU("digitalSignature", "nonRepudiation")
		case "encryption":
			addKU("keyEncipherment", "keyAgreement")
		}
	}
	if len(ku) == 0 {
		addKU("digitalSignature")
	}

	var sb strings.Builder
	sb.WriteString("basicConstraints=CA:FALSE\n")
	sb.WriteString("keyUsage=critical," + strings.Join(ku, ",") + "\n")
	if len(eku) > 0 {
		sb.WriteString("extendedKeyUsage=" + strings.Join(eku, ",") + "\n")
	}
	return os.WriteFile(path, []byte(sb.String()), 0600)
}
