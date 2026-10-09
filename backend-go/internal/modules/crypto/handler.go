package crypto

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"backend-go/internal/exception"
	"backend-go/internal/middleware"
	"backend-go/internal/response"
)

// Handler 密码操作 HTTP 处理器。
//
// 职责：
//  1. 通过 MetaRegistry 聚合各领域子 Handler（CA / Cert / CSR / CRL / Key）的路由；
//  2. 承载通用密码操作（/crypto/operations/:operation_id）与任务（/tasks/:task_id）路由。
//
// v1.3 修复：
//   - 此前 RegisterRoutes 遗漏对各子 Handler 的调用，导致 /api/v1/cas 等路由 404；
//   - 现改为通过 MetaRegistry 统一装配（MetaRegistry 已正确创建各子 Service）。
type Handler struct {
	svc  *Service
	meta *MetaRegistry
}

// NewHandler 创建密码操作处理器。
//
// 复用 Service 已持有的 db / cfg，构造 MetaRegistry，
// 避免在 main.go 中重复装配各子 Service。
func NewHandler(svc *Service) *Handler {
	return &Handler{
		svc:  svc,
		meta: NewMetaRegistry(svc.db, svc.cfg),
	}
}

// RegisterRoutes 注册密码操作路由（受保护）。
//
// 路由清单：
//
//	CA 管理：       /api/v1/cas/*
//	证书管理：      /api/v1/certs/*
//	CSR 管理：      /api/v1/csrs/*
//	CRL 管理：      /api/v1/crls/*
//	密钥管理：      /api/v1/keys/*
//	通用密码操作：   /api/v1/crypto/operations/:operation_id
//	任务：          /api/v1/tasks/:task_id
//	                /api/v1/tasks/:task_id/cancel
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

// GetTask 查询任务状态。
func (h *Handler) GetTask(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		response.Error(c, http.StatusBadRequest,
			exception.CodeParamInvalid, "缺少 task_id", nil)
		return
	}
	// 注意：实际实现中应调用 TaskScheduler 查询任务
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
