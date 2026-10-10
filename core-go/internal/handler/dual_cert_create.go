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

// DualCertCreateHandler dual_cert.create 国密双证签发。
//
// ★ 正确流程：
//   1. 客户本地生成 CSR，私钥 s_pri 永不上传（签名私钥归客户）；
//   2. core 用 CSR 的公钥 s_pub 签发【签名证书】（不重新生成签名密钥对）；
//   3. core 内部生成加密密钥对 e_pri / e_pub，用 e_pub 签发【加密证书】；
//   4. 返回 sign_cert_path / enc_cert_path / enc_key_ref（一次性）；
//      backend-go 侧拿到 enc_key_ref 后导出私钥 → 白盒加密 → 落盘 → 删除 core 引用。
//
// ★ 路径约定（无 -chain.pem）：
//   - sign_output_dir 非空：
//       <sign_output_dir>/<sign_pubkey_sm3>.cert.pem
//     sign_output_dir 为空：
//       <CoreRoot>/data/certs/<sign_cert_id>.pem
//   - enc_output_dir 非空：
//       <enc_output_dir>/<enc_pubkey_sm3>.cert.pem
//     enc_output_dir 为空：
//       <CoreRoot>/data/certs/<enc_cert_id>.pem
//
// ★ 证书校验：
//   依赖 <CertRoot>/ca/<domain>/ 下的 rehash 文件（<hash>.0），
//   由 openssl verify -CApath 自动定位，不需要 chain 文件。
type DualCertCreateHandler struct{}

func (h *DualCertCreateHandler) Execute(
	ctx context.Context, req *envelope.Request, cfg *config.Config,
) *envelope.Response {
	start := time.Now()

	// ---------- 算法校验 ----------
	signAlg := strParamDefault(req.Params, "sign_algorithm", "SM2")
	encAlg := strParamDefault(req.Params, "enc_algorithm", "SM2")
	tlcp := strParamDefault(req.Params, "tlcp_profile", "GB/T 38636-2020")

	if signAlg != "SM2" || encAlg != "SM2" || tlcp != "GB/T 38636-2020" {
		return envelope.NewErrorWithDetail("ALGORITHM_NOT_ALLOWED",
			"dual cert requires SM2+SM2+GB/T 38636-2020",
			map[string]interface{}{
				"sign_algorithm": signAlg,
				"enc_algorithm":  encAlg,
				"tlcp_profile":   tlcp,
			})
	}

	// ---------- CSR（必填） ----------
	csrPath := strParam(req.Params, "csr_path")
	if csrPath == "" {
		return genErr("INVALID_PARAM", "missing csr_path (客户 CSR 必须提供)")
	}

	// ---------- 父 CA ----------
	caCertRel := strParam(req.Params, "ca_cert_path")
	caKeyRef := strParam(req.Params, "ca_key_ref")
	if caCertRel == "" || caKeyRef == "" {
		return genErr("INVALID_PARAM", "missing ca_cert_path or ca_key_ref")
	}

	days := intParam(req.Params, "validity_days", 365)
	if days <= 0 {
		days = 365
	}

	// ---------- 输出路径参数 ----------
	signOutputDir := strParam(req.Params, "sign_output_dir")
	encOutputDir := strParam(req.Params, "enc_output_dir")
	outputLayout := strParamDefault(req.Params, "output_layout", "server_pubkey_sm3")

	// ---------- 解析 CSR ----------
	csrAbs, err := pathguard.Resolve(csrPath)
	if err != nil {
		return genErr("PATH_NOT_ALLOWED", "invalid csr_path: "+err.Error())
	}
	if _, err := os.Stat(csrAbs); err != nil {
		return genErr("INVALID_PARAM", "csr file not found: "+csrPath)
	}

	// ---------- 解析父 CA ----------
	caCertAbs, err := pathguard.Resolve(caCertRel)
	if err != nil {
		return genErr("PATH_NOT_ALLOWED", err.Error())
	}
	if _, err := os.Stat(caCertAbs); err != nil {
		return genErr("CERT_NOT_FOUND", "ca cert not found")
	}

	// ---------- 临时目录 ----------
	tmpDir, err := mkTmpDir("dual-cert")
	if err != nil {
		return genErr("CORE_EXEC_FAILED", err.Error())
	}
	defer cleanupTmpDir(tmpDir)

	// ---------- 解密父 CA 私钥 ----------
	caKeyAbs, err := loadDecryptedKey(caKeyRef, tmpDir)
	if err != nil {
		return genErr("KEY_NOT_FOUND", err.Error())
	}

	client, err := openssl.NewClient()
	if err != nil {
		return genErr("CORE_EXEC_FAILED", err.Error())
	}

	digest := digestForCertAlg(ctx, client, caCertAbs)

	// ========================================================================
	// 步骤 0：确定使用者 DN
	//
	//   优先 params.subject（用户显式输入）；
	//   否则从 CSR 中提取。
	//
	//   userProvidedSubject=true 时：
	//     - 签名证书：通过 openssl x509 -subj 覆盖 CSR 里的 DN
	//     - 加密证书：用该 DN 生成加密 CSR
	//   userProvidedSubject=false 时：
	//     - 签名证书：直接用 CSR，不加 -subj
	//     - 加密证书：用从 CSR 提取的 DN
	// ========================================================================
	userProvidedSubject := false
	subjectArg := ""
	if subject, ok := subjectFromParams(req.Params); ok && len(subject) > 0 {
		subjectArg = subjectToOpenSSL(subject)
		userProvidedSubject = true
	}
	if subjectArg == "" {
		subjOut, err := client.Run(ctx, "req", "-in", csrAbs, "-noout", "-subject", "-nameopt", "RFC2253")
		if err == nil {
			subjStr := strings.TrimSpace(string(subjOut))
			subjStr = strings.TrimSpace(strings.TrimPrefix(subjStr, "subject="))
			if subjStr != "" {
				subjectArg = rfc2253ToSubj(subjStr)
			}
		}
	}
	if subjectArg == "" {
		return genErr("INVALID_PARAM", "cannot determine subject (既未提供 subject，也无法从 CSR 提取)")
	}

	// ========================================================================
	// 步骤 1：用 CSR 签发【签名证书】
	//
	// ★ 关键：openssl x509 -req 会从 CSR 中读取公钥并写入证书，
	//   因此签名证书公钥 = CSR 公钥 = 客户 s_pub。
	//   客户私钥 s_pri 永远不上传，core 也不落签名私钥。
	// ========================================================================
	signCertID := newID("dual-sign")
	var (
		signCertAbs      string
		signCertRespPath string
		signUseNew       bool
	)

	if signOutputDir != "" {
		signUseNew = true
		if !filepath.IsAbs(signOutputDir) {
			return genErr("INVALID_PARAM", "sign_output_dir must be absolute")
		}
		if err := os.MkdirAll(signOutputDir, 0750); err != nil {
			return genErr("CORE_EXEC_FAILED", "mkdir sign_output_dir: "+err.Error())
		}
		signCertAbs = filepath.Join(signOutputDir, signCertID+".cert.pem.tmp")
	} else {
		signCertRel := "data/certs/" + signCertID + ".pem"
		var rerr error
		signCertAbs, rerr = pathguard.Resolve(signCertRel)
		if rerr != nil {
			return genErr("PATH_NOT_ALLOWED", rerr.Error())
		}
		if err := pathguard.EnsureDir(filepath.Dir(signCertAbs), 0750); err != nil {
			return genErr("CORE_EXEC_FAILED", err.Error())
		}
		signCertRespPath = signCertRel
	}

	// 签名证书扩展配置
	signExt := filepath.Join(tmpDir, "sign.ext")
	signExtContent := "basicConstraints=CA:FALSE\n" +
		"keyUsage=critical,digitalSignature,nonRepudiation\n" +
		"extendedKeyUsage=clientAuth\n"
	if err := os.WriteFile(signExt, []byte(signExtContent), 0600); err != nil {
		return genErr("CORE_EXEC_FAILED", err.Error())
	}

	signArgs := []string{
		"x509", "-req",
		"-in", csrAbs,
		"-CA", caCertAbs,
		"-CAkey", caKeyAbs,
		"-CAcreateserial",
		"-out", signCertAbs,
		"-days", fmt.Sprintf("%d", days),
		"-extfile", signExt,
	}
	if userProvidedSubject {
		signArgs = append(signArgs, "-subj", subjectArg)
	}
	if digest != "" {
		signArgs = append(signArgs, digest)
	}
	if _, err := client.Run(ctx, signArgs...); err != nil {
		audit.Log(req, "dual_cert.create", start, "FAILED", "CORE_EXEC_FAILED")
		return genErr("CORE_EXEC_FAILED", "sign cert: "+err.Error())
	}
	_ = os.Chmod(signCertAbs, 0640)

	if _, err := os.Stat(signCertAbs); err != nil {
		return genErr("CORE_EXEC_FAILED", "sign cert not generated")
	}

	// ★ 新规范：算 sign_pubkey_sm3，重命名
	var signPubSM3 string
	if signUseNew {
		signPubSM3, err = pubkeySM3FromCertAbs(ctx, client, signCertAbs, tmpDir)
		if err != nil {
			_ = os.Remove(signCertAbs)
			return genErr("CORE_EXEC_FAILED", "calc sign pubkey_sm3: "+err.Error())
		}
		finalSignCertAbs := filepath.Join(signOutputDir, signPubSM3+".cert.pem")
		if err := os.Rename(signCertAbs, finalSignCertAbs); err != nil {
			_ = os.Remove(signCertAbs)
			return genErr("CORE_EXEC_FAILED", "rename sign cert: "+err.Error())
		}
		signCertAbs = finalSignCertAbs
		signCertRespPath = signCertAbs
	}

	// ========================================================================
	// 步骤 2：内部生成加密密钥对（e_pri / e_pub）
	// ========================================================================
	encKeyPath := filepath.Join(tmpDir, "enc.key")
	if err := genPrivateKey(ctx, "SM2", req.Params, encKeyPath); err != nil {
		return genErr("CORE_EXEC_FAILED", "gen enc key: "+err.Error())
	}

	// 加密证书使用与签名证书相同的 DN
	encCSR := filepath.Join(tmpDir, "enc.csr")
	if _, err := client.Run(ctx, "req", "-new", "-key", encKeyPath,
		"-out", encCSR, "-subj", subjectArg); err != nil {
		return genErr("CORE_EXEC_FAILED", "enc csr: "+err.Error())
	}

	// ========================================================================
	// 步骤 3：用父 CA 签发【加密证书】
	// ========================================================================
	encCertID := newID("dual-enc")
	var (
		encCertAbs      string
		encCertRespPath string
		encUseNew       bool
	)

	if encOutputDir != "" {
		encUseNew = true
		if !filepath.IsAbs(encOutputDir) {
			return genErr("INVALID_PARAM", "enc_output_dir must be absolute")
		}
		if err := os.MkdirAll(encOutputDir, 0750); err != nil {
			return genErr("CORE_EXEC_FAILED", "mkdir enc_output_dir: "+err.Error())
		}
		encCertAbs = filepath.Join(encOutputDir, encCertID+".cert.pem.tmp")
	} else {
		encCertRel := "data/certs/" + encCertID + ".pem"
		var rerr error
		encCertAbs, rerr = pathguard.Resolve(encCertRel)
		if rerr != nil {
			return genErr("PATH_NOT_ALLOWED", rerr.Error())
		}
		if err := pathguard.EnsureDir(filepath.Dir(encCertAbs), 0750); err != nil {
			return genErr("CORE_EXEC_FAILED", err.Error())
		}
		encCertRespPath = encCertRel
	}

	encExt := filepath.Join(tmpDir, "enc.ext")
	encExtContent := "basicConstraints=CA:FALSE\n" +
		"keyUsage=critical,keyEncipherment,keyAgreement\n" +
		"extendedKeyUsage=clientAuth\n"
	if err := os.WriteFile(encExt, []byte(encExtContent), 0600); err != nil {
		return genErr("CORE_EXEC_FAILED", err.Error())
	}

	encArgs := []string{
		"x509", "-req",
		"-in", encCSR,
		"-CA", caCertAbs,
		"-CAkey", caKeyAbs,
		"-CAcreateserial",
		"-out", encCertAbs,
		"-days", fmt.Sprintf("%d", days),
		"-extfile", encExt,
	}
	if digest != "" {
		encArgs = append(encArgs, digest)
	}
	if _, err := client.Run(ctx, encArgs...); err != nil {
		audit.Log(req, "dual_cert.create", start, "FAILED", "CORE_EXEC_FAILED")
		return genErr("CORE_EXEC_FAILED", "enc cert: "+err.Error())
	}
	_ = os.Chmod(encCertAbs, 0640)

	if _, err := os.Stat(encCertAbs); err != nil {
		return genErr("CORE_EXEC_FAILED", "enc cert not generated")
	}

	// ★ 新规范：算 enc_pubkey_sm3，重命名
	var encPubSM3 string
	if encUseNew {
		encPubSM3, err = pubkeySM3FromCertAbs(ctx, client, encCertAbs, tmpDir)
		if err != nil {
			_ = os.Remove(encCertAbs)
			return genErr("CORE_EXEC_FAILED", "calc enc pubkey_sm3: "+err.Error())
		}
		finalEncCertAbs := filepath.Join(encOutputDir, encPubSM3+".cert.pem")
		if err := os.Rename(encCertAbs, finalEncCertAbs); err != nil {
			_ = os.Remove(encCertAbs)
			return genErr("CORE_EXEC_FAILED", "rename enc cert: "+err.Error())
		}
		encCertAbs = finalEncCertAbs
		encCertRespPath = encCertAbs
	}

	// ========================================================================
	// 步骤 4：加密私钥存入 keystore，返回一次性 key_ref
	//
	// ★ 加密私钥不下落到 output_dir：
	//   - core 侧：keystore 加密存储，返回 enc_key_ref；
	//   - backend-go 侧：拿 enc_key_ref 调 key.manage export 导出明文，
	//     再调 whitebox_sm4 白盒加密，落盘到 <enc_dir>/<enc_pub>.key.pem，
	//     最后调 key.manage delete 删除 core 侧临时引用。
	// ========================================================================
	encPlain, err := os.ReadFile(encKeyPath)
	if err != nil {
		return genErr("CORE_EXEC_FAILED", "read enc key: "+err.Error())
	}
	encKeyRef, err := saveEncryptedKey(encPlain)
	if err != nil {
		return genErr("CORE_EXEC_FAILED", "save enc key: "+err.Error())
	}

	// ========================================================================
	// 步骤 5：审计 + 响应（★ 无 chain_path）
	// ========================================================================
	audit.Log(req, "dual_cert.create", start, "SUCCESS", "")

	data := map[string]interface{}{
		"sign_cert_path": signCertRespPath,
		"sign_cert_id":   signCertID,
		"enc_cert_path":  encCertRespPath,
		"enc_cert_id":    encCertID,
		"enc_key_ref":    encKeyRef,
	}
	if signUseNew {
		data["sign_pubkey_sm3"] = signPubSM3
	}
	if encUseNew {
		data["enc_pubkey_sm3"] = encPubSM3
	}
	data["output_layout"] = outputLayout

	return envelope.NewSuccess(data)
}

// =============================================================================
// RFC2253 → openssl -subj 风格转换
// =============================================================================

// rfc2253ToSubj 把 RFC2253 subject 转换为 openssl -subj 风格。
//
// 输入： "CN=Test,O=Example,C=CN"
// 输出： "/C=CN/O=Example/CN=Test"
func rfc2253ToSubj(rfc string) string {
	parts := splitRFC2253(rfc)
	kv := make(map[string]string)
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		idx := strings.Index(p, "=")
		if idx <= 0 {
			continue
		}
		k := strings.ToUpper(strings.TrimSpace(p[:idx]))
		v := strings.TrimSpace(p[idx+1:])
		kv[k] = v
	}
	order := []string{"C", "ST", "L", "O", "OU", "CN", "E", "SERIALNUMBER"}
	var sb strings.Builder
	used := map[string]bool{}
	for _, k := range order {
		if v, ok := kv[k]; ok && v != "" {
			sb.WriteString("/")
			sb.WriteString(k)
			sb.WriteString("=")
			sb.WriteString(v)
			used[k] = true
		}
	}
	for k, v := range kv {
		if used[k] || v == "" {
			continue
		}
		sb.WriteString("/")
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(v)
	}
	return sb.String()
}

// splitRFC2253 按未转义的逗号拆分 RFC2253 字符串。
func splitRFC2253(s string) []string {
	var (
		parts   []string
		current strings.Builder
		escaped bool
	)
	for _, r := range s {
		switch {
		case escaped:
			current.WriteRune(r)
			escaped = false
		case r == '\\':
			current.WriteRune(r)
			escaped = true
		case r == ',':
			parts = append(parts, current.String())
			current.Reset()
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}
