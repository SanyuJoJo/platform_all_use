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

// CSRCreateHandler csr.create。
//
// ★ 证书路径改造：
//   - output_dir 非空时走新规范：
//       <output_dir>/<pubkey_sm3>.req.csr
//   - output_dir 为空时保持旧行为：
//       <CoreRoot>/data/csr/<csr_id>.csr
type CSRCreateHandler struct{}

func (h *CSRCreateHandler) Execute(
	ctx context.Context, req *envelope.Request, cfg *config.Config,
) *envelope.Response {
	start := time.Now()

	wl, _ := config.GetWhitelist(cfg.WhitelistPath)
	alg := strParamDefault(req.Params, "algorithm", "SM2")
	if wl != nil && !wl.IsAlgorithmAllowed("csr.create", alg) {
		return genErr("ALGORITHM_NOT_ALLOWED", "algorithm not allowed")
	}

	subject, ok := subjectFromParams(req.Params)
	if !ok {
		return genErr("INVALID_PARAM", "missing subject")
	}

	// ★ 输出路径参数
	outputDir := strParam(req.Params, "output_dir")
	outputLayout := strParamDefault(req.Params, "output_layout", "server_pubkey_sm3")

	tmpDir, err := mkTmpDir("csr-create")
	if err != nil {
		return genErr("CORE_EXEC_FAILED", err.Error())
	}
	defer cleanupTmpDir(tmpDir)

	keyPath := filepath.Join(tmpDir, "key.pem")
	if err := genPrivateKey(ctx, alg, req.Params, keyPath); err != nil {
		return genErr("CORE_EXEC_FAILED", "gen key: "+err.Error())
	}

	plainKey, _ := os.ReadFile(keyPath)
	keyRef, err := saveEncryptedKey(plainKey)
	if err != nil {
		return genErr("CORE_EXEC_FAILED", err.Error())
	}

	// ★ 输出路径决策
	csrID := newID("csr")
	var (
		csrAbs       string
		csrRespPath  string
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
		csrAbs = filepath.Join(outputDir, csrID+".req.csr.tmp")
	} else {
		csrRel := "data/csr/" + csrID + ".csr"
		var rerr error
		csrAbs, rerr = pathguard.Resolve(csrRel)
		if rerr != nil {
			return genErr("PATH_NOT_ALLOWED", rerr.Error())
		}
		if err := pathguard.EnsureDir(filepath.Dir(csrAbs), 0750); err != nil {
			return genErr("CORE_EXEC_FAILED", err.Error())
		}
		csrRespPath = csrRel
	}

	// SAN
	var sanLine string
	if raw, ok := req.Params["san"].([]interface{}); ok && len(raw) > 0 {
		parts := make([]string, 0, len(raw))
		for _, v := range raw {
			if s, ok := v.(string); ok && s != "" {
				parts = append(parts, "DNS:"+s)
			}
		}
		if len(parts) > 0 {
			sanLine = "subjectAltName=" + strings.Join(parts, ",") + "\n"
		}
	}

	// 生成扩展配置
	extPath := filepath.Join(tmpDir, "csr.ext")
	var eb strings.Builder
	eb.WriteString("[req]\ndistinguished_name=dn\nreq_extensions=v3_req\nprompt=no\n")
	eb.WriteString("[dn]\n")
	for k, v := range subject {
		eb.WriteString(fmt.Sprintf("%s=%s\n", k, v))
	}
	eb.WriteString("[v3_req]\nbasicConstraints=CA:FALSE\n")
	if sanLine != "" {
		eb.WriteString(sanLine)
	}
	_ = os.WriteFile(extPath, []byte(eb.String()), 0600)

	client, err := openssl.NewClient()
	if err != nil {
		return genErr("CORE_EXEC_FAILED", err.Error())
	}

	if _, err := client.Run(ctx, "req", "-new", "-key", keyPath,
		"-out", csrAbs, "-config", extPath); err != nil {
		return genErr("CORE_EXEC_FAILED", "csr: "+err.Error())
	}
	_ = os.Chmod(csrAbs, 0640)

	// ★ 新规范：算 pubkey_sm3，重命名为 <pubkey_sm3>.req.csr
	var pubKeySM3 string
	if useNewLayout {
		pubKeySM3, err = pubkeySM3FromCSRAbs(ctx, client, csrAbs, tmpDir)
		if err != nil {
			_ = os.Remove(csrAbs)
			audit.Log(req, "csr.create", start, "FAILED", "CORE_EXEC_FAILED")
			return genErr("CORE_EXEC_FAILED", "calc pubkey_sm3: "+err.Error())
		}
		finalCSRAbs := filepath.Join(outputDir, pubKeySM3+".req.csr")
		if err := os.Rename(csrAbs, finalCSRAbs); err != nil {
			_ = os.Remove(csrAbs)
			return genErr("CORE_EXEC_FAILED", "rename csr: "+err.Error())
		}
		csrAbs = finalCSRAbs
		csrRespPath = csrAbs
	}

	audit.Log(req, "csr.create", start, "SUCCESS", "")

	data := map[string]interface{}{
		"csr_id":   csrID,
		"csr_path": csrRespPath,
		"key_ref":  keyRef,
	}
	if useNewLayout {
		data["pubkey_sm3"] = pubKeySM3
		data["output_layout"] = outputLayout
	}
	return envelope.NewSuccess(data)
}
