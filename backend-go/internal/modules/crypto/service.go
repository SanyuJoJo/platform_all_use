package crypto

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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
	caller     *CoreCaller
	whitebox   *WhiteboxClient
	cfg        *config.Config
	auditSvc   *audit_log.Service
	db         *gorm.DB
	layout     *PathLayout
	sem        chan struct{}
	taskMu     sync.Mutex
	taskCancel map[string]context.CancelFunc
}

// NewService 创建密码操作服务。
func NewService(cfg *config.Config, auditSvc *audit_log.Service, db *gorm.DB) *Service {
	coreBin := strings.TrimSpace(cfg.CoreGoBin)
	if coreBin == "" {
		coreBin = strings.TrimSpace(cfg.CoreDispatchPath)
	}
	if coreBin == "" && cfg.InstallRoot != "" {
		coreBin = filepath.Join(cfg.InstallRoot, "core", "bin", "core")
	}

	coreRoot := strings.TrimSpace(cfg.CoreRoot)
	if coreRoot == "" && cfg.InstallRoot != "" {
		coreRoot = filepath.Join(cfg.InstallRoot, "core")
	}

	adapter := NewCoreAdapter(coreBin, cfg.CoreTimeoutMs, cfg.CoreMaxConcurrency)
	layout := NewPathLayout(cfg)
	_ = layout.EnsureRoots()

	files := NewFileStore(coreRoot)
	caller := NewCoreCaller(adapter, coreRoot, files, layout)
	whitebox := NewWhiteboxClient(cfg.CoreGoWhiteboxBin)

	return &Service{
		adapter:    adapter,
		caller:     caller,
		whitebox:   whitebox,
		cfg:        cfg,
		auditSvc:   auditSvc,
		db:         db,
		layout:     layout,
		sem:        make(chan struct{}, cfg.CoreMaxConcurrency),
		taskCancel: make(map[string]context.CancelFunc),
	}
}

// HealthCheck 健康检查。
func (s *Service) HealthCheck() error {
	return s.adapter.HealthCheck()
}

// Execute 执行密码操作（同步）。
//
// ★ 关键点：构造 coreReq.Params 时使用**副本**并调用 applyCoreAlgorithmMapping，
// 把 params.algorithm 从平台 ID（RSA2048/MLDSA65/...）转成 core 名（RSA/ML-DSA-65/...）。
// persistMetadata 仍用 req.Params（平台 ID）落库。
func (s *Service) Execute(
	ctx context.Context,
	op string,
	req *OperationRequest,
	actorType, actorID string,
) (*OperationResponse, int, error) {
	if !IsValidOperation(op) {
		return nil, 400, exception.New(
			exception.CodeParamInvalid,
			fmt.Sprintf("未知的 operation_id: %s", op),
			400, nil,
		)
	}

	requestID := uuid.NewString()

	// 参数预校验
	if err := ValidateSubjectForOperation(op, req.Params); err != nil {
		msg := err.Error()
		platformCode := PlatformCryptoInvalidParam
		if pe, ok := err.(*exception.PlatformError); ok {
			msg = pe.Message
		}
		s.writeAudit(requestID, op, "FAILED", 0, platformCode, nil)
		return &OperationResponse{
			Code:        platformCode,
			Message:     msg,
			RequestID:   requestID,
			OperationID: op,
			Error: &PlatformErrorDetail{
				Code:      platformCode,
				Message:   msg,
				Retryable: false,
			},
		}, 400, nil
	}

	// 证书路径改造：在通用路由上注入 output_dir
	if req.Params == nil {
		req.Params = map[string]interface{}{}
	}
	if err := InjectOutputLayout(s.layout, op, req.Params); err != nil {
		msg := err.Error()
		platformCode := PlatformCryptoInvalidParam
		if pe, ok := err.(*exception.PlatformError); ok {
			msg = pe.Message
		}
		s.writeAudit(requestID, op, "FAILED", 0, platformCode, nil)
		return &OperationResponse{
			Code:        platformCode,
			Message:     msg,
			RequestID:   requestID,
			OperationID: op,
			Error: &PlatformErrorDetail{
				Code:      platformCode,
				Message:   msg,
				Retryable: false,
			},
		}, 400, nil
	}

	// ★ 构造 core 请求：用副本，把 algorithm 映射为 core 名
	coreParams := cloneParams(req.Params)
	applyCoreAlgorithmMapping(coreParams)

	// 构造 core 请求
	coreReq := &CoreRequest{
		SchemaVersion: "1.0",
		OperationID:   op,
		RequestID:     requestID,
		Actor: CoreActor{
			Type: actorType,
			ID:   actorID,
		},
		Params: coreParams,
	}
	if req.Options != nil {
		coreReq.Options = &CoreOptions{
			TimeoutMs: req.Options.TimeoutMs,
			DryRun:    req.Options.DryRun,
		}
	}

	// 并发信号量
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

	// 调用 CoreAdapter
	start := time.Now()
	coreResp, exitCode, err := s.adapter.Call(ctx, op, coreReq)
	durationMs := int(time.Since(start).Milliseconds())

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

	mapping := MapCoreError(coreResp.Code)

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

		// 元数据落库（用 req.Params 里的平台 ID）
		s.persistMetadata(op, req.Params, coreResp.Data)

		// ★ csr.create 后处理：把 core 私钥导出并白盒加密落盘
		if op == "csr.create" {
			s.postProcessCsrCreate(ctx, req.Params, coreResp.Data)
		}
	}

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

// postProcessCsrCreate 把 core 生成的 CSR 私钥导出并白盒加密落盘。
func (s *Service) postProcessCsrCreate(
	ctx context.Context, params, data map[string]interface{},
) {
	if s.caller == nil || s.whitebox == nil {
		return
	}

	if kp := getString(data, "key_path"); kp != "" {
		abs := kp
		if !filepath.IsAbs(abs) && s.layout != nil && s.layout.CertRoot != "" {
			abs = filepath.Join(s.layout.CertRoot, abs)
		}
		if st, err := os.Stat(abs); err == nil && !st.IsDir() && st.Size() > 0 {
			log.Debug().Str("path", abs).
				Msg("postProcessCsrCreate: core 已落盘 key.pem，跳过")
			return
		}
	}

	keyRef := getString(data, "key_ref")
	if keyRef == "" {
		log.Debug().Msg("postProcessCsrCreate: 无 key_ref，跳过")
		return
	}

	csrPath := getString(data, "csr_path")
	if csrPath == "" {
		if csrID := getString(data, "csr_id"); csrID != "" && s.db != nil {
			var csr models.CSR
			if err := s.db.Where("csr_id = ?", csrID).First(&csr).Error; err == nil {
				csrPath = csr.CSRPath
			}
		}
	}
	if csrPath == "" {
		log.Warn().Str("key_ref", keyRef).
			Msg("postProcessCsrCreate: 无法确定 csr_path，跳过")
		return
	}

	absCsr := s.resolveCsrAbs(csrPath)
	if absCsr == "" || !strings.HasSuffix(absCsr, ".req.csr") {
		log.Warn().Str("csr_path", csrPath).
			Msg("postProcessCsrCreate: csr_path 不符合规范，跳过")
		return
	}
	keyAbs := strings.TrimSuffix(absCsr, ".req.csr") + ".key.pem"

	if st, err := os.Stat(keyAbs); err == nil && !st.IsDir() && st.Size() > 0 {
		log.Debug().Str("path", keyAbs).
			Msg("postProcessCsrCreate: key.pem 已存在，跳过")
		return
	}

	plain, err := s.caller.ExportKey(ctx, keyRef)
	if err != nil {
		log.Error().Err(err).Str("key_ref", keyRef).
			Msg("postProcessCsrCreate: 导出私钥失败（忽略）")
		return
	}
	if len(plain) == 0 {
		log.Warn().Str("key_ref", keyRef).
			Msg("postProcessCsrCreate: 导出的私钥为空（忽略）")
		return
	}
	defer func() {
		for i := range plain {
			plain[i] = 0
		}
	}()

	tmpDir, err := os.MkdirTemp("", "csr-key-")
	if err != nil {
		log.Error().Err(err).Msg("postProcessCsrCreate: 创建临时目录失败")
		return
	}
	defer os.RemoveAll(tmpDir)

	plainTmp := filepath.Join(tmpDir, "plain.key")
	if err := os.WriteFile(plainTmp, plain, 0600); err != nil {
		log.Error().Err(err).Msg("postProcessCsrCreate: 写临时明文失败")
		return
	}
	if err := s.whitebox.EncryptFile(ctx, plainTmp, keyAbs); err != nil {
		log.Error().Err(err).Str("path", keyAbs).
			Msg("postProcessCsrCreate: 白盒加密失败（忽略）")
		_ = os.Remove(keyAbs)
		return
	}
	_ = os.Chmod(keyAbs, 0600)

	log.Info().
		Str("csr_path", absCsr).
		Str("key_path", keyAbs).
		Str("key_ref", keyRef).
		Msg("postProcessCsrCreate: CSR 私钥已白盒加密落盘")
}

// resolveCsrAbs 解析 CSR 相对/绝对路径为存在的绝对路径。
func (s *Service) resolveCsrAbs(p string) string {
	if p == "" {
		return ""
	}
	if filepath.IsAbs(p) {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	if s.layout != nil && s.layout.CertRoot != "" {
		abs := filepath.Join(s.layout.CertRoot, p)
		if _, err := os.Stat(abs); err == nil {
			return abs
		}
	}
	if s.caller != nil {
		abs := s.caller.ResolveCorePath(p)
		if _, err := os.Stat(abs); err == nil {
			return abs
		}
	}
	if filepath.IsAbs(p) {
		return p
	}
	return ""
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

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func intPtr(v int) *int {
	return &v
}

// =============================================================================
// 辅助
// =============================================================================

// cloneParams 浅拷贝 params，用于传给 core 前做算法名映射（避免污染调用方）。
func cloneParams(src map[string]interface{}) map[string]interface{} {
	if src == nil {
		return map[string]interface{}{}
	}
	dst := make(map[string]interface{}, len(src)+2)
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
