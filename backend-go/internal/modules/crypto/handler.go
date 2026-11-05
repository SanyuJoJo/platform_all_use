// backend-go/internal/modules/crypto/handler.go
package crypto

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"backend-go/internal/exception"
	"backend-go/internal/middleware"
	"backend-go/internal/response"
)

// Handler 密码操作 HTTP 处理器。
//
// 职责：
//  1. 通过 MetaRegistry 聚合各领域子 Handler（CA / Cert / CSR / CRL / Key）的路由；
//  2. 承载通用密码操作（/crypto/operations/:operation_id）与任务（/tasks/:task_id）路由；
//  3. ★ 拦截 csr.create + p10_format=cfca，改走 cfca_envelope 生成 CFCA P10。
type Handler struct {
	svc  *Service
	meta *MetaRegistry
}

// NewHandler 创建密码操作处理器。
func NewHandler(svc *Service) *Handler {
	return &Handler{
		svc:  svc,
		meta: NewMetaRegistry(svc.db, svc.cfg),
	}
}

// Meta 返回内部 MetaRegistry，供其它模块复用 CA/Cert Service。
func (h *Handler) Meta() *MetaRegistry {
	return h.meta
}

// RegisterRoutes 注册密码操作路由（受保护）。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// ---- 各领域子路由（通过 MetaRegistry 统一注册）----
	h.meta.Register(r)

	// ---- 通用密码操作路由 ----
	crypto := r.Group("/crypto")
	{
		crypto.POST("/operations/:operation_id",
			middleware.RequirePermission("crypto:operation:execute"),
			h.Execute,
		)
	}

	// ---- 任务路由 ----
	tasks := r.Group("/tasks")
	{
		tasks.GET("/:task_id",
			middleware.RequirePermission("crypto:task:view"),
			h.GetTask,
		)
		tasks.POST("/:task_id/cancel",
			middleware.RequirePermission("crypto:task:cancel"),
			h.CancelTask,
		)
	}
}

// Execute 执行密码操作。
func (h *Handler) Execute(c *gin.Context) {
	op := c.Param("operation_id")
	if op == "" {
		response.Error(c, http.StatusBadRequest,
			exception.CodeParamInvalid, "缺少 operation_id", nil)
		return
	}

	var req OperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "请求参数校验失败",
			gin.H{"errors": []map[string]interface{}{
				{"type": "invalid_json", "msg": err.Error()},
			}})
		return
	}
	if req.Params == nil {
		req.Params = map[string]interface{}{}
	}

	// ★ CFCA P10 分支：p10_format=cfca 时走 cfca_envelope gen-csr，
	//   而不是 core 的 csr.create（core 只生成普通 PKCS10，不会携带
	//   临时公钥 attribute）。
	if op == "csr.create" && isCFCAP10Format(req.Params) {
		h.handleCreateCFCACSR(c, &req)
		return
	}

	// 获取当前用户信息
	actorType := "platform-backend"
	actorID := ""
	if val, ok := c.Get(middleware.ContextKeyUserID); ok {
		if uid, ok := val.(uint); ok {
			actorID = fmt.Sprintf("%d", uid)
		}
	}
	if val, ok := c.Get(middleware.ContextKeyUserInfo); ok {
		if uc, ok := val.(middleware.UserContext); ok {
			actorType = "platform-user"
			actorID = uc.Username
		}
	}

	// 执行
	resp, httpStatus, err := h.svc.Execute(c.Request.Context(), op, &req, actorType, actorID)
	if err != nil {
		handleError(c, err)
		return
	}

	// 根据任务是否异步返回不同状态码
	if resp.TaskID != nil {
		response.SuccessWithStatus(c, http.StatusAccepted, resp, "任务已提交")
		return
	}
	response.SuccessWithStatus(c, httpStatus, resp, resp.Message)
}

// =============================================================================
// ★ CFCA P10 分支
// =============================================================================

// isCFCAP10Format 判断请求是否指定了 p10_format=cfca。
func isCFCAP10Format(params map[string]interface{}) bool {
	v, ok := params["p10_format"]
	if !ok || v == nil {
		return false
	}
	if s, ok := v.(string); ok {
		return strings.EqualFold(strings.TrimSpace(s), "cfca")
	}
	return false
}

// extractSubjectFromParams 从 params["subject"] 提取 subject。
//
// 兼容两种类型（JSON 反序列化后为 map[string]interface{}，
// 手工构造时可能为 map[string]string）。
func extractSubjectFromParams(params map[string]interface{}) map[string]string {
	raw, ok := params["subject"]
	if !ok || raw == nil {
		return nil
	}
	switch m := raw.(type) {
	case map[string]string:
		out := make(map[string]string, len(m))
		for k, v := range m {
			out[k] = v
		}
		return out
	case map[string]interface{}:
		out := make(map[string]string, len(m))
		for k, v := range m {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				out[k] = s
			}
		}
		return out
	}
	return nil
}

// extractSANFromParams 从 params["san"] 提取 SAN 列表。
func extractSANFromParams(params map[string]interface{}) []string {
	raw, ok := params["san"]
	if !ok || raw == nil {
		return nil
	}
	switch v := raw.(type) {
	case []string:
		out := make([]string, 0, len(v))
		for _, s := range v {
			if t := strings.TrimSpace(s); t != "" {
				out = append(out, t)
			}
		}
		return out
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				if t := strings.TrimSpace(s); t != "" {
					out = append(out, t)
				}
			}
		}
		return out
	}
	return nil
}

// handleCreateCFCACSR 处理 "申请 P10" 的 CFCA 分支。
//
// 流程：
//  1. CFCAService.GenerateCSRForApply  → 生成携带临时公钥 attribute 的 CSR + 签名私钥；
//  2. CSRService.Import               → 落盘 CSR/私钥、落库 models.CSR；
//  3. 组装契约响应（csr_id / csr_pem / csr_path / pubkey_sm3 / sign_key_pem / is_cfca）。
//
// 之所以复用 CSRService.Import 而不是自己写落盘 + 落库：Import 已处理
// dir 分配、pubSM3 命名、白盒加密、fillCSRDerivedFields（含 is_cfca 判定）
// 等全部细节，逻辑等价且不需要重复。
func (h *Handler) handleCreateCFCACSR(c *gin.Context, req *OperationRequest) {
	subject := extractSubjectFromParams(req.Params)
	if len(subject) == 0 || strings.TrimSpace(subject["CN"]) == "" {
		response.Error(c, http.StatusBadRequest,
			exception.CodeParamInvalid, "CFCA P10 需要 subject.CN", nil)
		return
	}

	// CFCA 仅支持 SM2
	if alg := strings.TrimSpace(getString(req.Params, "algorithm")); alg != "" &&
		!strings.EqualFold(alg, "SM2") {
		response.Error(c, http.StatusBadRequest,
			exception.CodeParamInvalid,
			"CFCA P10 仅支持 SM2 算法", nil)
		return
	}

	san := extractSANFromParams(req.Params)

	ctx := c.Request.Context()

	// 1. 生成 CSR + 签名私钥（走 cfca_envelope gen-csr --tmp-key <内置私钥>）
	result, err := h.meta.CFCA.GenerateCSRForApply(ctx, subject, san)
	if err != nil {
		handleError(c, err)
		return
	}

	// 2. 落盘 + 落库（CSRService.Import 内部写 .req.csr、白盒加密 .key.pem，
	//    并调用 fillCSRDerivedFields → is_cfca=true）
	importReq := &ImportCsrRequest{
		CSRPEM: string(result.CSRPEM),
		KeyPEM: string(result.SignKeyPEM),
	}
	csr, err := h.meta.CSR.Import(ctx, importReq)
	if err != nil {
		handleError(c, err)
		return
	}

	// 3. 组装契约响应
	//    pubkey_sm3 从 csr_path 文件名推导（<pubSM3>.req.csr），与 core 普通
	//    分支返回的同名字段保持一致的语义。
	pubSM3 := ""
	if csr.CSRPath != "" {
		if base := filepath.Base(csr.CSRPath); strings.HasSuffix(base, ".req.csr") {
			pubSM3 = strings.TrimSuffix(base, ".req.csr")
		}
	}

	resp := map[string]interface{}{
		"csr_id":               csr.CSRID,
		"csr_pem":              string(result.CSRPEM),
		"csr_path":             csr.CSRPath,
		"pubkey_sm3":           pubSM3,
		"sign_key_pem":         string(result.SignKeyPEM),
		"has_private_key":      csr.HasPrivateKey,
		"is_cfca":              csr.IsCFCA,
		"subject_cn":           csr.SubjectCN,
		"public_key_algorithm": csr.Algorithm,
		"status":               csr.Status,
		"created_at":           csr.CreatedAt,
		"not_before":           csr.NotBefore,
	}

	response.SuccessWithStatus(c, http.StatusCreated, resp, "CFCA P10 申请成功")
}

// GetTask 查询任务状态。
func (h *Handler) GetTask(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		response.Error(c, http.StatusBadRequest,
			exception.CodeParamInvalid, "缺少 task_id", nil)
		return
	}
	response.Success(c, gin.H{
		"task_id": taskID,
		"status":  "RUNNING",
	}, "success")
}

// CancelTask 取消任务。
func (h *Handler) CancelTask(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		response.Error(c, http.StatusBadRequest,
			exception.CodeParamInvalid, "缺少 task_id", nil)
		return
	}
	if err := h.svc.CancelTask(taskID); err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, nil, "任务已取消")
}

// handleError 统一处理 PlatformError。
func handleError(c *gin.Context, err error) {
	var pe *exception.PlatformError
	if errors.As(err, &pe) {
		response.Error(c, pe.HTTPStatus, pe.Code, pe.Message, pe.Data)
		return
	}
	response.Error(c, http.StatusInternalServerError,
		exception.CodeInternalError, "服务器内部错误", nil)
}
