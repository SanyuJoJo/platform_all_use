package crypto

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"backend-go/internal/exception"
)

// CoreCaller 封装 CoreAdapter，提供领域无关的 core 调用能力。
//
// 任何子 Service（CAService / CertService / ...）都通过 CoreCaller
// 与 core 交互，避免重复实现错误码映射、request_id 生成、响应校验。
type CoreCaller struct {
	adapter  *CoreAdapter
	coreRoot string
	files    *FileStore
}

// NewCoreCaller 创建 CoreCaller。
func NewCoreCaller(adapter *CoreAdapter, coreRoot string, files *FileStore) *CoreCaller {
	return &CoreCaller{adapter: adapter, coreRoot: coreRoot, files: files}
}

// CoreRoot 返回 core 根目录（供子 Service 拼路径用）。
func (c *CoreCaller) CoreRoot() string { return c.coreRoot }

// Run 调用任意 operation_id，统一做错误码映射与响应校验。
func (c *CoreCaller) Run(
	ctx context.Context, op string, params map[string]interface{},
) (*CoreResponse, error) {
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

// ParseCert 调用 cert.parse，返回解析后的结构化字段。
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

// ImportKey 调用 key.manage import，返回 key_ref。
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

// ExportKey 导出私钥到受控临时路径，返回内容并自动清理。
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

// ExportPKCS12 导出 PKCS#12 到受控临时路径，返回内容并自动清理。
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
