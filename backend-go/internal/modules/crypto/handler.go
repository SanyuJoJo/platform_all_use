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
type Handler struct {
	svc *Service
}

// NewHandler 创建密码操作处理器。
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes 注册密码操作路由（受保护）。
//
// 路由：
//   POST /api/v1/crypto/operations/:operation_id  执行密码操作
//   GET  /api/v1/tasks/:task_id                    查询任务状态
//   POST /api/v1/tasks/:task_id/cancel             取消任务
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	crypto := r.Group("/crypto")
	{
		crypto.POST("/operations/:operation_id",
			middleware.RequirePermission("crypto:operation:execute"),
			h.Execute,
		)
	}
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
	// 此处为示例
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
