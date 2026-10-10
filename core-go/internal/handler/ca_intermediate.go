package handler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/yourorg/core-go/internal/audit"
	"github.com/yourorg/core-go/internal/config"
	"github.com/yourorg/core-go/internal/envelope"
	"github.com/yourorg/core-go/internal/openssl"
	"github.com/yourorg/core-go/internal/pathguard"
)

// CAIntermediateHandler ca.intermediate.create 中间 CA 生成。
//
// ★ 路径约定（无 chain）：
//   - output_dir 非空：
//       <output_dir>/<pubkey_sm3>.cert.pem
//       （CA 私钥走 keystore，返回 key_ref）
//   - output_dir 为空：data/ca/<ca_id>.pem
type CAIntermediateHandler struct{}

func (h *CAIntermediateHandler) Execute(
	ctx context.Context, req *envelope.Request, cfg *config.Config,
) *envelope.Response {
	start := time.Now()

	// ---------- 白名单 ----------
	wl, err := config.GetWhitelist(cfg.WhitelistPath)
	if err != nil {
		return genErr("INTERNAL_ERROR", "whitelist load failed")
	}
	alg := strParam(req.Params, "algorithm")
	if !wl.IsAlgorithmAllowed("ca.intermediate.create", alg) {
		return genErr("ALGORITHM_NOT_ALLOWED", "algorithm not allowed")
	}

	// ---------- 父 CA ----------
	parentCertRel := strParam(req.Params, "parent_ca_cert_path")
	parentKeyRef := strParam(req.Params, "parent_ca_key_ref")
	if parentCertRel == "" || parentKeyRef == "" {
		return genErr("INVALID_PARAM", "missing parent_ca_cert_path or parent_ca_key_ref")
	}

	parentCertAbs, err := pathguard.Resolve(parentCertRel)
	if err != nil {
		return genErr("PATH_NOT_ALLOWED", err.Error())
	}
	if _, err := os.Stat(parentCertAbs); err != nil {
		return genErr("CERT_NOT_FOUND", "parent ca cert not found")
	}

	// ---------- subject / validity ----------
	subject, ok := subjectFromParams(req.Params)
	if !ok {
		return genErr("INVALID_PARAM", "missing subject")
	}
	days := intParam(req.Params, "validity_days", 1825)
	pathLen := intParam(req.Params, "path_len", 0)

	// ---------- 输出路径参数 ----------
	outputDir := strParam(req.Params, "output_dir")
	outputLayout := strParamDefault(req.Params, "output_layout", "ca_domain_pubkey_sm3")

	// ---------- 临时目录 ----------
	tmpDir, err := mkTmpDir("ca-inter")
	if err != nil {
		return genErr("CORE_EXEC_FAILED", err.Error())
	}
	defer cleanupTmpDir(tmpDir)

	// ---------- 解密父 CA 私钥 ----------
	parentKeyAbs, err := loadDecryptedKey(parentKeyRef, tmpDir)
	if err != nil {
		return genErr("KEY_NOT_FOUND", err.Error())
	}

	// ---------- 生成中间 CA 私钥 ----------
	keyPath := filepath.Join(tmpDir, "key.pem")
	if err := genPrivateKey(ctx, alg, req.Params, keyPath); err != nil {
		return genErr("CORE_EXEC_FAILED", "gen key: "+err.Error())
	}

	plainKey, _ := os.ReadFile(keyPath)
	keyRef, err := saveEncryptedKey(plainKey)
	if err != nil {
		return genErr("CORE_EXEC_FAILED", err.Error())
	}

	// ---------- 输出路径决策 ----------
	caID := newID("ca-int")
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
		certAbs = filepath.Join(outputDir, caID+".cert.pem.tmp")
	} else {
		certRel := "data/ca/" + caID + ".pem"
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

	// ---------- 生成 CSR ----------
	csrPath := filepath.Join(tmpDir, "req.csr")
	subjArg := subjectToOpenSSL(subject)

	client, err := openssl.NewClient()
	if err != nil {
		return genErr("CORE_EXEC_FAILED", err.Error())
	}

	if _, err := client.Run(ctx, "req", "-new", "-key", keyPath,
		"-out", csrPath, "-subj", subjArg); err != nil {
		return genErr("CORE_EXEC_FAILED", "csr: "+err.Error())
	}

	// ---------- 扩展配置 ----------
	extPath := filepath.Join(tmpDir, "ca.ext")
	extContent := fmt.Sprintf(
		"basicConstraints=critical,CA:TRUE,pathlen:%d\n"+
			"keyUsage=critical,keyCertSign,cRLSign\n",
		pathLen)
	if err := os.WriteFile(extPath, []byte(extContent), 0600); err != nil {
		return genErr("CORE_EXEC_FAILED", err.Error())
	}

	// ---------- 签发 ----------
	digest := digestForCertAlg(ctx, client, parentCertAbs)

	if _, err := client.Run(ctx, "x509", "-req", "-in", csrPath,
		"-CA", parentCertAbs, "-CAkey", parentKeyAbs, "-CAcreateserial",
		"-out", certAbs, "-days", fmt.Sprintf("%d", days),
		digest, "-extfile", extPath); err != nil {
		return genErr("CORE_EXEC_FAILED", "sign: "+err.Error())
	}
	_ = os.Chmod(certAbs, 0640)

	// ---------- 新规范：重命名为 <pubkey_sm3>.cert.pem ----------
	var pubKeySM3 string
	if useNewLayout {
		pubKeySM3, err = pubkeySM3FromCertAbs(ctx, client, certAbs, tmpDir)
		if err != nil {
			_ = os.Remove(certAbs)
			audit.Log(req, "ca.intermediate.create", start, "FAILED", "CORE_EXEC_FAILED")
			return genErr("CORE_EXEC_FAILED", "calc pubkey_sm3: "+err.Error())
		}
		finalCertAbs := filepath.Join(outputDir, pubKeySM3+".cert.pem")
		if err := os.Rename(certAbs, finalCertAbs); err != nil {
			_ = os.Remove(certAbs)
			return genErr("CORE_EXEC_FAILED", "rename cert: "+err.Error())
		}
		certAbs = finalCertAbs
		certRespPath = certAbs

		// ★ rehash 软链：<pubkey_sm3>.0 → <pubkey_sm3>.cert.pem
		// 使该证书可作为 CA 校验（openssl -CApath 依赖 <hash>.0 命名）
		if err := writeRehashLink(certAbs); err != nil {
			// 失败仅告警，不阻断
			audit.Log(req, "ca.intermediate.create", start, "SUCCESS",
				"rehash failed: "+err.Error())
		}
	}

	// ---------- 序列号 ----------
	serial := extractSerial(ctx, client, certAbs)

	// ---------- 审计 + 响应（★ 无 chain_path） ----------
	audit.Log(req, "ca.intermediate.create", start, "SUCCESS", "")

	data := map[string]interface{}{
		"ca_id":     caID,
		"cert_path": certRespPath,
		"key_ref":   keyRef,
		"serial":    serial,
	}
	if useNewLayout {
		data["pubkey_sm3"] = pubKeySM3
		data["output_layout"] = outputLayout
	}
	return envelope.NewSuccess(data)
}
