package handler

import (
	"context"
	"os"
	"path/filepath"

	"github.com/yourorg/core-go/internal/config"
	"github.com/yourorg/core-go/internal/envelope"
	"github.com/yourorg/core-go/internal/openssl"
	"github.com/yourorg/core-go/internal/pathguard"
)

// KeyManageHandler key.manage 密钥管理。
//
// 支持 4 种 action：
//   - generate ：生成密钥对，加密存 keystore，返回 key_ref
//   - import   ：导入外部私钥，加密存 keystore，返回 key_ref
//   - export   ：导出私钥明文（需 allow_plain_export=true）
//   - delete   ：软删除 keystore 中的私钥（重命名为 .deleted）
type KeyManageHandler struct{}

func (h *KeyManageHandler) Execute(
	ctx context.Context, req *envelope.Request, cfg *config.Config,
) *envelope.Response {
	action := strParam(req.Params, "action")
	switch action {
	case "generate":
		return h.generate(ctx, req)
	case "import":
		return h.importKey(ctx, req)
	case "export":
		return h.exportKey(ctx, req)
	case "delete":
		return h.deleteKey(req)
	default:
		return genErr("INVALID_PARAM", "invalid action: "+action)
	}
}

// -----------------------------------------------------------------------------
// generate
// -----------------------------------------------------------------------------

// generate 生成密钥对并加密存 keystore。
func (h *KeyManageHandler) generate(
	ctx context.Context, req *envelope.Request,
) *envelope.Response {
	alg := strParamDefault(req.Params, "algorithm", "SM2")

	tmpDir, err := mkTmpDir("keygen")
	if err != nil {
		return genErr("CORE_EXEC_FAILED", err.Error())
	}
	defer cleanupTmpDir(tmpDir)

	keyPath := filepath.Join(tmpDir, "key.pem")
	if err := genPrivateKey(ctx, alg, req.Params, keyPath); err != nil {
		return genErr("INVALID_PARAM", err.Error())
	}

	plain, err := os.ReadFile(keyPath)
	if err != nil {
		return genErr("CORE_EXEC_FAILED", "read key: "+err.Error())
	}
	keyRef, err := saveEncryptedKey(plain)
	if err != nil {
		return genErr("CORE_EXEC_FAILED", "save key: "+err.Error())
	}

	return envelope.NewSuccess(map[string]interface{}{
		"key_ref":   keyRef,
		"algorithm": alg,
		"state":     "ACTIVE",
	})
}

// -----------------------------------------------------------------------------
// import
// -----------------------------------------------------------------------------

// importKey 导入外部私钥。
//
// ★ 加固点：
//   - 用 `pkey -in <key> -noout -passin pass:` 校验；
//     显式提供空密码，openssl 不会从 tty 读取；
//   - 加密私钥会因空密码失败 → 立即返回错误，不挂起；
//   - 明文私钥空密码可读 → 成功，继续导入。
//
// 调用方（backend-go）**必须**先解密，把明文私钥传入。
func (h *KeyManageHandler) importKey(
	ctx context.Context, req *envelope.Request,
) *envelope.Response {
	keyRel := strParam(req.Params, "key_path")
	if keyRel == "" {
		return genErr("INVALID_PARAM", "missing key_path")
	}

	abs, err := pathguard.Resolve(keyRel)
	if err != nil {
		return genErr("PATH_NOT_ALLOWED", err.Error())
	}
	if _, err := os.Stat(abs); err != nil {
		return genErr("KEY_NOT_FOUND", "key file not found")
	}

	client, err := openssl.NewClient()
	if err != nil {
		return genErr("CORE_EXEC_FAILED", err.Error())
	}

	// ★ 校验私钥有效性：显式空密码 → 加密私钥会立即失败，不挂起
	if _, err := client.Run(ctx, "pkey",
		"-in", abs,
		"-noout",
		"-passin", "pass:",
	); err != nil {
		return genErr("CERT_PARSE_FAILED",
			"invalid private key (or encrypted, please decrypt first): "+err.Error())
	}

	// 读取明文并加密存 keystore
	plain, err := os.ReadFile(abs)
	if err != nil {
		return genErr("CORE_EXEC_FAILED", "read key: "+err.Error())
	}
	keyRef, err := saveEncryptedKey(plain)
	if err != nil {
		return genErr("CORE_EXEC_FAILED", "save key: "+err.Error())
	}

	return envelope.NewSuccess(map[string]interface{}{
		"key_ref": keyRef,
	})
}

// -----------------------------------------------------------------------------
// export
// -----------------------------------------------------------------------------

// exportKey 导出私钥明文到受控路径。
//
// ★ 必须显式传 allow_plain_export=true，否则拒绝。
func (h *KeyManageHandler) exportKey(
	ctx context.Context, req *envelope.Request,
) *envelope.Response {
	keyRef := strParam(req.Params, "key_ref")
	if keyRef == "" {
		return genErr("INVALID_PARAM", "missing key_ref")
	}

	exportRel := strParamDefault(req.Params, "export_path",
		"data/export/"+keyRef+".pem")

	allow := false
	if v, ok := req.Params["allow_plain_export"].(bool); ok {
		allow = v
	}
	if !allow {
		return genErr("PERMISSION_DENIED", "plain export disabled")
	}

	tmpDir, err := mkTmpDir("keyexport")
	if err != nil {
		return genErr("CORE_EXEC_FAILED", err.Error())
	}
	defer cleanupTmpDir(tmpDir)

	plainPath, err := loadDecryptedKey(keyRef, tmpDir)
	if err != nil {
		return genErr("KEY_NOT_FOUND", err.Error())
	}

	exportAbs, err := pathguard.Resolve(exportRel)
	if err != nil {
		return genErr("PATH_NOT_ALLOWED", err.Error())
	}
	if err := pathguard.EnsureDir(filepath.Dir(exportAbs), 0750); err != nil {
		return genErr("CORE_EXEC_FAILED", err.Error())
	}

	data, err := os.ReadFile(plainPath)
	if err != nil {
		return genErr("CORE_EXEC_FAILED", "read plain key: "+err.Error())
	}
	if err := os.WriteFile(exportAbs, data, 0600); err != nil {
		return genErr("CORE_EXEC_FAILED", "write export: "+err.Error())
	}

	return envelope.NewSuccess(map[string]interface{}{
		"key_ref":     keyRef,
		"export_path": exportRel,
	})
}

// -----------------------------------------------------------------------------
// delete
// -----------------------------------------------------------------------------

// deleteKey 软删除 keystore 中的私钥。
//
// 将 <keysDir>/<keyRef>.key.enc 重命名为 <keyRef>.key.enc.deleted。
// 不物理删除，保留可追溯性。
func (h *KeyManageHandler) deleteKey(req *envelope.Request) *envelope.Response {
	keyRef := strParam(req.Params, "key_ref")
	if keyRef == "" {
		return genErr("INVALID_PARAM", "missing key_ref")
	}

	keysDir := pathguard.MustResolve("data/keys")
	src := filepath.Join(keysDir, keyRef+".key.enc")
	dst := src + ".deleted"

	if err := os.Rename(src, dst); err != nil {
		return genErr("KEY_NOT_FOUND", err.Error())
	}

	return envelope.NewSuccess(map[string]interface{}{
		"key_ref": keyRef,
		"state":   "DELETED",
	})
}
