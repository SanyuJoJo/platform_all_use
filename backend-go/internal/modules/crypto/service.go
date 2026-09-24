package crypto

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"backend-go/internal/config"
	"backend-go/internal/exception"
	"backend-go/internal/models"
	"backend-go/internal/modules/audit_log"
)

// Service 密码操作服务。
type Service struct {
	adapter    *CoreAdapter
	cfg        *config.Config
	auditSvc   *audit_log.Service
	db         *gorm.DB // ← 新增：用于元数据落库
	sem        chan struct{}
	taskMu     sync.Mutex
	taskCancel map[string]context.CancelFunc
}

// NewService 创建密码操作服务。
//
// 变更：新增 db 参数，用于把 core 返回的元数据写入平台表。
func NewService(cfg *config.Config, auditSvc *audit_log.Service, db *gorm.DB) *Service {
	adapter := NewCoreAdapter(
		cfg.CoreDispatchPath,
		cfg.CoreTimeoutMs,
		cfg.CoreMaxConcurrency,
	)
	return &Service{
		adapter:    adapter,
		cfg:        cfg,
		auditSvc:   auditSvc,
		db:         db,
		sem:        make(chan struct{}, cfg.CoreMaxConcurrency),
		taskCancel: make(map[string]context.CancelFunc),
	}
}

// HealthCheck 健康检查。
func (s *Service) HealthCheck() error {
	return s.adapter.HealthCheck()
}

// Execute 执行密码操作（同步）。
func (s *Service) Execute(
	ctx context.Context,
	op string,
	req *OperationRequest,
	actorType, actorID string,
) (*OperationResponse, int, error) {
	// 1. 校验 operation_id
	if !IsValidOperation(op) {
		return nil, 400, exception.New(
			exception.CodeParamInvalid,
			fmt.Sprintf("未知的 operation_id: %s", op),
			400, nil,
		)
	}

	// 2. 生成 request_id
	requestID := uuid.NewString()

	// 3. 构造 core 请求
	coreReq := &CoreRequest{
		SchemaVersion: "1.0",
		OperationID:   op,
		RequestID:     requestID,
		Actor: CoreActor{
			Type: actorType,
			ID:   actorID,
		},
		Params: req.Params,
	}
	if req.Options != nil {
		coreReq.Options = &CoreOptions{
			TimeoutMs: req.Options.TimeoutMs,
			DryRun:    req.Options.DryRun,
		}
	}

	// 4. 并发信号量
	select {
	case s.sem <- struct{}{}:
		defer func() { <-s.sem }()
	case <-ctx.Done():
		return nil, 504, exception.New(
			exception.CodeInternalError,
			"请求已取消",
			504, nil,
		)
	}

	// 5. 调用 CoreAdapter
	start := time.Now()
	coreResp, exitCode, err := s.adapter.Call(ctx, op, coreReq)
	durationMs := int(time.Since(start).Milliseconds())

	// 6. 处理调用错误
	if err != nil {
		platformCode := PlatformCryptoCoreFailed
		httpStatus := 500
		if strings.Contains(err.Error(), "超时") {
			platformCode = PlatformCryptoCoreTimeout
			httpStatus = 504
		}
		s.writeAudit(requestID, op, "FAILED", durationMs, platformCode, nil)
		return &OperationResponse{
			Code:        platformCode,
			Message:     err.Error(),
			RequestID:   requestID,
			OperationID: op,
			Error: &PlatformErrorDetail{
				Code:      platformCode,
				Message:   err.Error(),
				Retryable: false,
			},
		}, httpStatus, nil
	}

	// 7. 映射错误码
	mapping := MapCoreError(coreResp.Code)

	// 8. 构造平台响应
	platformResp := &OperationResponse{
		Code:        mapping.PlatformCode,
		Message:     coreResp.Message,
		RequestID:   coreResp.RequestID,
		OperationID: coreResp.OperationID,
		TaskID:      coreResp.TaskID,
		Data:        coreResp.Data,
		Audit: &PlatformAudit{
			AuditID:      requestID,
			ParamsDigest: "",
			Result:       "SUCCESS",
			DurationMs:   durationMs,
		},
	}

	if coreResp.Code != "OK" {
		platformResp.Error = &PlatformErrorDetail{
			Code:      mapping.PlatformCode,
			Message:   coreResp.Message,
			Detail:    coreResp.Error.Detail,
			Retryable: mapping.Retryable,
		}
		if coreResp.Audit != nil {
			platformResp.Audit.ParamsDigest = coreResp.Audit.ParamsDigest
			platformResp.Audit.Result = coreResp.Audit.Result
		}
		s.writeAudit(requestID, op, "FAILED", durationMs, mapping.PlatformCode, coreResp.Error.Detail)
	} else {
		if coreResp.Audit != nil {
			platformResp.Audit.ParamsDigest = coreResp.Audit.ParamsDigest
			platformResp.Audit.Result = coreResp.Audit.Result
		}
		s.writeAudit(requestID, op, "SUCCESS", durationMs, "", nil)

		// ★★★ 关键新增：core 成功后，把元数据落库 ★★★
		s.persistMetadata(op, req.Params, coreResp.Data)
	}

	// 9. 调试日志
	log.Debug().
		Str("operation_id", op).
		Str("request_id", requestID).
		Int("exit_code", exitCode).
		Str("core_code", coreResp.Code).
		Str("platform_code", mapping.PlatformCode).
		Int("duration_ms", durationMs).
		Msg("密码操作完成")

	return platformResp, mapping.HTTPStatus, nil
}

// ExecuteAsync 执行密码操作（异步，返回 task_id）。
func (s *Service) ExecuteAsync(
	ctx context.Context,
	op string,
	req *OperationRequest,
	actorType, actorID string,
) (*OperationResponse, error) {
	if !IsValidOperation(op) {
		return nil, exception.New(
			exception.CodeParamInvalid,
			fmt.Sprintf("未知的 operation_id: %s", op),
			400, nil,
		)
	}

	taskID := uuid.NewString()
	requestID := uuid.NewString()

	taskCtx, cancel := context.WithCancel(context.Background())
	s.taskMu.Lock()
	s.taskCancel[taskID] = cancel
	s.taskMu.Unlock()

	go func() {
		defer func() {
			s.taskMu.Lock()
			delete(s.taskCancel, taskID)
			s.taskMu.Unlock()
			cancel()
		}()

		resp, _, err := s.Execute(taskCtx, op, req, actorType, actorID)
		if err != nil {
			log.Error().Err(err).Str("task_id", taskID).Msg("异步任务失败")
			return
		}
		if resp.Code == PlatformSuccess {
			log.Info().Str("task_id", taskID).Msg("异步任务成功")
		} else {
			log.Warn().
				Str("task_id", taskID).
				Str("code", resp.Code).
				Msg("异步任务失败")
		}
	}()

	return &OperationResponse{
		Code:        PlatformSuccess,
		Message:     "任务已提交",
		RequestID:   requestID,
		OperationID: op,
		TaskID:      &taskID,
		Data: map[string]interface{}{
			"task_id": taskID,
		},
	}, nil
}

// CancelTask 取消任务。
func (s *Service) CancelTask(taskID string) error {
	s.taskMu.Lock()
	cancel, ok := s.taskCancel[taskID]
	s.taskMu.Unlock()
	if !ok {
		return exception.New(
			exception.CodeParamInvalid,
			"任务不存在或已结束",
			404, nil,
		)
	}
	cancel()
	return nil
}

// writeAudit 写平台审计。
func (s *Service) writeAudit(
	requestID, operationID, result string,
	durationMs int, errorCode string, detail map[string]interface{},
) {
	if s.auditSvc == nil {
		return
	}
	entry := &models.AuditLogOperation{
		ModuleID:  "crypto",
		Action:    operationID,
		Detail:    strPtr(fmt.Sprintf("operation=%s", operationID)),
		Status:    strings.ToLower(result),
		RequestID: strPtr(requestID),
	}
	if errorCode != "" {
		entry.ErrorCode = intPtr(90000)
	}
	s.auditSvc.WriteOperationLogAsync(entry)
}

// 辅助函数
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func intPtr(v int) *int {
	return &v
}
