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

// CACreateHandler ca.create 根 CA 生成。
//
// ★ 证书路径改造：
//   - output_dir 非空时走新规范：
//       <output_dir>/<pubkey_sm3>.cert.pem
//       <output_dir>/<pubkey_sm3>.key.pem       白盒密文（密钥内嵌二进制）
//       <output_dir>/<hash>.0                   rehash 软链（★ 本次补上）
//     响应 cert_path 为绝对路径；同时返回 pubkey_sm3 / key_path。
//   - output_dir 为空时保持旧行为：
//       <CoreRoot>/data/ca/<ca_id>.pem
//     私钥走 keystore，响应返回 key_ref。
type CACreateHandler struct{}

func (h *CACreateHandler) Execute(
	ctx context.Context, req *envelope.Request, cfg *config.Config,
) *envelope.Response {
	start := time.Now()

	// ---------- 白名单 ----------
	wl, err := config.GetWhitelist(cfg.WhitelistPath)
	if err != nil {
		return genErr("INTERNAL_ERROR", "whitelist load failed: "+err.Error())
	}
	alg := strParam(req.Params, "algorithm")
	if !wl.IsAlgorithmAllowed("ca.create", alg) {
		audit.Log(req, "ca.create", start, "FAILED", "ALGORITHM_NOT_ALLOWED")
		return envelope.NewErrorWithDetail("ALGORITHM_NOT_ALLOWED",
			"algorithm not allowed",
			map[string]interface{}{"algorithm": alg})
	}

	// ---------- subject / validity ----------
	subject, ok := subjectFromParams(req.Params)
	if !ok {
		audit.Log(req, "ca.create", start, "FAILED", "INVALID_PARAM")
		return genErr("INVALID_PARAM", "missing subject")
	}
	validityDays := intParam(req.Params, "validity_days", 3650)
	if validityDays <= 0 {
		validityDays = 3650
	}

	// ---------- 路径参数 ----------
	outputDir := strParam(req.Params, "output_dir")
	outputLayout := strParamDefault(req.Params, "output_layout", "ca_domain_pubkey_sm3")
	domain := strParam(req.Params, "domain")

	// ---------- 临时目录 ----------
	tmpDir, err := mkTmpDir("ca-create")
	if err != nil {
		return genErr("CORE_EXEC_FAILED", err.Error())
	}
	defer cleanupTmpDir(tmpDir)

	// ---------- 生成私钥 ----------
	keyPath := filepath.Join(tmpDir, "key.pem")
	if err := genPrivateKey(ctx, alg, req.Params, keyPath); err != nil {
		audit.Log(req, "ca.create", start, "FAILED", "CORE_EXEC_FAILED")
		return genErr("CORE_EXEC_FAILED", "gen key: "+err.Error())
	}
	plainKey, err := os.ReadFile(keyPath)
	if err != nil {
		return genErr("CORE_EXEC_FAILED", err.Error())
	}

	// ---------- 落盘路径决策 ----------
	caID := newID("ca")
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

	// ---------- 生成 CA 证书 ----------
	subjArg := subjectToOpenSSL(subject)
	digest := digestForAlgorithm(alg)

	client, err := openssl.NewClient()
	if err != nil {
		return genErr("CORE_EXEC_FAILED", err.Error())
	}

	args := []string{
		"req", "-new", "-x509",
		"-key", keyPath,
		"-out", certAbs,
		"-days", fmt.Sprintf("%d", validityDays),
		"-subj", subjArg,
		"-addext", "basicConstraints=critical,CA:TRUE",
		"-addext", "keyUsage=critical,keyCertSign,cRLSign",
	}
	if digest != "" {
		args = append(args, digest)
	}
	if _, err := client.Run(ctx, args...); err != nil {
		audit.Log(req, "ca.create", start, "FAILED", "CORE_EXEC_FAILED")
		return genErr("CORE_EXEC_FAILED", "openssl req: "+err.Error())
	}
	_ = os.Chmod(certAbs, 0640)

	// ---------- 返回变量 ----------
	var (
		pubKeySM3   string
		keyRef      string
		keyRespPath string
	)

	if useNewLayout {
		// ★ 新规范
		// 1. 算 pubkey_sm3
		pubKeySM3, err = pubkeySM3FromCertAbs(ctx, client, certAbs, tmpDir)
		if err != nil {
			_ = os.Remove(certAbs)
			audit.Log(req, "ca.create", start, "FAILED", "CORE_EXEC_FAILED")
			return genErr("CORE_EXEC_FAILED", "calc pubkey_sm3: "+err.Error())
		}

		// 2. 证书重命名为规范路径
		finalCertAbs := filepath.Join(outputDir, pubKeySM3+".cert.pem")
		if err := os.Rename(certAbs, finalCertAbs); err != nil {
			_ = os.Remove(certAbs)
			return genErr("CORE_EXEC_FAILED", "rename cert: "+err.Error())
		}
		certAbs = finalCertAbs
		certRespPath = certAbs

		// 3. 私钥白盒加密落盘
		keyAbs := filepath.Join(outputDir, pubKeySM3+".key.pem")
		if err := whiteboxEncrypt(ctx, cfg, keyPath, keyAbs); err != nil {
			_ = os.Remove(certAbs)
			audit.Log(req, "ca.create", start, "FAILED", "CORE_EXEC_FAILED")
			return genErr("CORE_EXEC_FAILED", "whitebox encrypt key: "+err.Error())
		}
		keyRespPath = keyAbs

		// 4. ★ 生成 rehash 软链 <subject_hash>.0
		//
		// 作用：让 openssl verify -CApath <outputDir> 能自动定位到本 CA。
		// 失败仅告警，不阻断创建（可后续用 c_rehash 修复）。
		if err := writeRehashLink(finalCertAbs); err != nil {
			audit.Log(req, "ca.create", start, "SUCCESS",
				"rehash failed: "+err.Error())
		}
	} else {
		// 旧规范：私钥走 keystore
		keyRef, err = saveEncryptedKey(plainKey)
		if err != nil {
			return genErr("CORE_EXEC_FAILED", "save key: "+err.Error())
		}
	}

	// ---------- 序列号 ----------
	serial := extractSerial(ctx, client, certAbs)

	audit.Log(req, "ca.create", start, "SUCCESS", "")

	// ---------- 响应 ----------
	data := map[string]interface{}{
		"ca_id":     caID,
		"cert_path": certRespPath,
		"serial":    serial,
	}
	if useNewLayout {
		data["pubkey_sm3"] = pubKeySM3
		data["key_path"] = keyRespPath
		data["output_layout"] = outputLayout
		if domain != "" {
			data["domain"] = domain
		}
	} else {
		data["key_ref"] = keyRef
	}

	return envelope.NewSuccess(data)
}
