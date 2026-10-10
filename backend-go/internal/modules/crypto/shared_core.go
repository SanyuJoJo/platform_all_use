package crypto

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"backend-go/internal/exception"
)

// CoreCaller 封装 CoreAdapter，提供领域无关的 core 调用能力。
type CoreCaller struct {
	adapter  *CoreAdapter
	coreRoot string
	files    *FileStore
	layout   *PathLayout
}

func NewCoreCaller(
	adapter *CoreAdapter, coreRoot string, files *FileStore, layout *PathLayout,
) *CoreCaller {
	return &CoreCaller{
		adapter:  adapter,
		coreRoot: coreRoot,
		files:    files,
		layout:   layout,
	}
}

func (c *CoreCaller) CoreRoot() string       { return c.coreRoot }
func (c *CoreCaller) Layout() *PathLayout     { return c.layout }

// Run 调用任意 operation_id。
func (c *CoreCaller) Run(
	ctx context.Context, op string, params map[string]interface{},
) (*CoreResponse, error) {
	if params == nil {
		params = map[string]interface{}{}
	}
	if err := InjectOutputLayout(c.layout, op, params); err != nil {
		return nil, err
	}

	req := &CoreRequest{
		SchemaVersion: "1.0",
		OperationID:   op,
		RequestID:     uuid.NewString(),
		Actor:         CoreActor{Type: "platform-backend", ID: "system"},
		Params:        params,
	}
	resp, _, err := c.adapter.Call(ctx, op, req)
	if err != nil {
		return nil, exception.New(
			exception.CodeInternalError,
			fmt.Sprintf("调用 core %s 失败: %v", op, err), 500, nil,
		)
	}
	if resp.Code != "OK" {
		m := MapCoreError(resp.Code)
		return nil, exception.New(
			exceptionCodeFromHTTP(m.HTTPStatus),
			fmt.Sprintf("[%s] %s", resp.Code, resp.Message),
			m.HTTPStatus, nil,
		)
	}
	return resp, nil
}

// =============================================================================
// 路径自动注入
// =============================================================================

// InjectOutputLayout 按 op 注入路径参数。
func InjectOutputLayout(
	layout *PathLayout, op string, params map[string]interface{},
) error {
	if layout == nil || params == nil {
		return nil
	}
	if _, exists := params["output_dir"]; exists {
		return nil
	}

	switch op {
	case "ca.create", "ca.intermediate.create":
		return injectCAOutput(layout, params)
	case "csr.create":
		return injectCSROutput(layout, params)
	case "cert.sign":
		return injectServerOutput(layout, params)
	case "dual_cert.create":
		return injectDualOutput(layout, params)
	}
	return nil
}

// injectCAOutput 为 ca.create 注入 CA 目录。
//
// ★ 变更：domain 默认值从 subject.CN 派生改为固定 "safe"。
//
// domain 来源优先级：
//  1. params["domain"]（显式指定）
//  2. "safe"（默认域）
func injectCAOutput(layout *PathLayout, params map[string]interface{}) error {
	domain := strings.TrimSpace(getString(params, "domain"))
	if domain == "" {
		domain = "safe" // ★ 默认域
	}

	dir, err := layout.CADir(domain)
	if err != nil {
		return exception.New(exception.CodeParamInvalid,
			fmt.Sprintf("非法 CA 域名目录：%v", err), 400, nil)
	}
	if err := os.MkdirAll(dir, 0750); err != nil {
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("创建 CA 目录失败：%v", err), 500, nil)
	}

	params["domain"] = domain
	params["output_dir"] = dir
	params["output_layout"] = "ca_domain_pubkey_sm3"
	params["cert_root"] = layout.CertRoot
	return nil
}

func injectCSROutput(layout *PathLayout, params map[string]interface{}) error {
	dirNo := strings.TrimSpace(getString(params, "dir_no"))
	if dirNo == "" {
		var err error
		dirNo, err = layout.AllocServerDir()
		if err != nil {
			return exception.New(exception.CodeInternalError,
				fmt.Sprintf("分配 server 目录失败：%v", err), 500, nil)
		}
	}
	dir, err := layout.ServerDir(dirNo)
	if err != nil {
		return exception.New(exception.CodeParamInvalid,
			fmt.Sprintf("server 目录非法：%v", err), 400, nil)
	}

	params["dir_no"] = dirNo
	params["output_dir"] = dir
	params["output_layout"] = "server_pubkey_sm3"
	params["cert_root"] = layout.CertRoot
	return nil
}

func injectServerOutput(layout *PathLayout, params map[string]interface{}) error {
	if _, ok := params["dir_no"]; ok {
		return nil
	}
	return injectCSROutput(layout, params)
}

func injectDualOutput(layout *PathLayout, params map[string]interface{}) error {
	if _, ok := params["sign_dir_no"]; !ok {
		signDirNo, err := layout.AllocServerDir()
		if err != nil {
			return exception.New(exception.CodeInternalError,
				fmt.Sprintf("分配签名 server 目录失败：%v", err), 500, nil)
		}
		signDir, _ := layout.ServerDir(signDirNo)
		params["sign_dir_no"] = signDirNo
		params["sign_output_dir"] = signDir
	}
	if _, ok := params["enc_dir_no"]; !ok {
		encDirNo, err := layout.AllocServerDir()
		if err != nil {
			return exception.New(exception.CodeInternalError,
				fmt.Sprintf("分配加密 server 目录失败：%v", err), 500, nil)
		}
		encDir, _ := layout.ServerDir(encDirNo)
		params["enc_dir_no"] = encDirNo
		params["enc_output_dir"] = encDir
	}

	params["output_layout"] = "server_pubkey_sm3"
	params["cert_root"] = layout.CertRoot
	return nil
}

// extractCNFromSubjectMap 保留（供其它调用者）。
func extractCNFromSubjectMap(v interface{}) string {
	m, ok := v.(map[string]interface{})
	if !ok {
		if m2, ok2 := v.(map[string]string); ok2 {
			if cn, ok := m2["CN"]; ok {
				return cn
			}
			if cn, ok := m2["cn"]; ok {
				return cn
			}
		}
		return ""
	}
	for _, k := range []string{"CN", "cn", "CommonName", "common_name"} {
		if s, ok := m[k].(string); ok && strings.TrimSpace(s) != "" {
			return s
		}
	}
	return ""
}

// ResolveCorePath 解析 core 返回的相对路径。
func (c *CoreCaller) ResolveCorePath(rel string) string {
	if rel == "" {
		return ""
	}
	if filepath.IsAbs(rel) {
		return rel
	}
	return filepath.Join(c.coreRoot, rel)
}

// =============================================================================
// 领域方法
// =============================================================================

func (c *CoreCaller) ParseCert(
	ctx context.Context, certRel string,
) (map[string]interface{}, error) {
	resp, err := c.Run(ctx, "cert.parse", map[string]interface{}{
		"cert_path": certRel,
	})
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *CoreCaller) ImportKey(
	ctx context.Context, keyRel, algorithm string,
) (string, error) {
	if algorithm == "" {
		algorithm = "SM2"
	}
	resp, err := c.Run(ctx, "key.manage", map[string]interface{}{
		"action":    "import",
		"key_path":  keyRel,
		"algorithm": algorithm,
	})
	if err != nil {
		return "", err
	}
	keyRef := getString(resp.Data, "key_ref")
	if keyRef == "" {
		return "", exception.New(
			exception.CodeInternalError, "key.manage 未返回 key_ref", 500, nil,
		)
	}
	return keyRef, nil
}

func (c *CoreCaller) ExportKey(ctx context.Context, keyRef string) ([]byte, error) {
	id := newShortID()
	exportRel := fmt.Sprintf("tmp/export-%s.key.pem", id)

	_, err := c.Run(ctx, "key.manage", map[string]interface{}{
		"action":             "export",
		"key_ref":            keyRef,
		"export_path":        exportRel,
		"allow_plain_export": true,
	})
	if err != nil {
		return nil, err
	}
	data, err := c.files.ReadCoreFile(exportRel)
	_ = c.files.SafeRemove(exportRel)
	return data, err
}

func (c *CoreCaller) ExportPKCS12(
	ctx context.Context, certRel, keyRef, password string,
) ([]byte, error) {
	id := newShortID()
	pwRel := fmt.Sprintf("tmp/export-%s.pw", id)
	targetRel := fmt.Sprintf("tmp/export-%s.p12", id)

	if err := c.files.WriteCoreFile(pwRel, []byte(password), 0600); err != nil {
		return nil, err
	}
	defer c.files.SafeRemove(pwRel)

	_, err := c.Run(ctx, "cert.convert", map[string]interface{}{
		"source_format": "PEM",
		"target_format": "PKCS12",
		"source_path":   certRel,
		"target_path":   targetRel,
		"password_file": pwRel,
		"key_ref":       keyRef,
	})
	if err != nil {
		return nil, err
	}
	data, err := c.files.ReadCoreFile(targetRel)
	_ = c.files.SafeRemove(targetRel)
	return data, err
}
