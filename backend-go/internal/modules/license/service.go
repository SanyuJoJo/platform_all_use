package license
import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"backend-go/internal/config"
	"backend-go/internal/exception"
	"backend-go/internal/models"
	"backend-go/internal/modules/auth"
)
// Service License 管理服务。
type Service struct {
	db         *gorm.DB
	cfg        *config.Config
	httpClient *http.Client
}
// NewService 创建 License 管理服务。
func NewService(db *gorm.DB, cfg *config.Config) *Service {
	return &Service{
		db:         db,
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}
// ---------------------------------------------------------------------------
// 内部工具
// ---------------------------------------------------------------------------
func utcNowNaive() time.Time {
	return time.Now().UTC()
}
// ptr 泛型辅助函数，返回 v 的指针（Go-P3-03）。
func ptr[T any](v T) *T {
	return &v
}
// normalizeAuthorizedModules 保证 nil slice 归一化为空数组（Go-P1-01）。
//
// 与 Python 版的 `list(license_obj.authorized_modules or [])` 行为一致：
//   - DB 存储 NULL 时，models.StringArray 为 nil，直接序列化为 JSON null；
//   - 归一化后统一返回 []string{}，序列化为 JSON []。
//
// 该函数是 v1.1 修复 P1-01 的核心，必须同时在
// serializeLicenseForImport 与 GetLicenseStatus 中调用。
func normalizeAuthorizedModules(a models.StringArray) []string {
	if a == nil {
		return []string{}
	}
	return []string(a)
}
// parseISODateTime 解析 ISO 8601 字符串为 UTC 时间。
func parseISODateTime(value interface{}, field string) (time.Time, error) {
	s, ok := value.(string)
	if !ok || s == "" {
		return time.Time{}, exception.New(
			exception.CodeLicenseInvalidFile,
			fmt.Sprintf("License 字段 %s 必须为时间字符串", field),
			400, nil,
		)
	}
	// 规范化：Z 后缀 → +00:00
	normalized := s
	if strings.HasSuffix(normalized, "Z") {
		normalized = strings.TrimSuffix(normalized, "Z") + "+00:00"
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05.000000",
		"2006-01-02 15:04:05.000000",
		"2006-01-02T15:04:05.000",
		"2006-01-02 15:04:05.000",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, normalized); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, exception.New(
		exception.CodeLicenseInvalidFile,
		fmt.Sprintf("License 字段 %s 时间格式无效：%s", field, s),
		400, nil,
	)
}
// loadActiveLicense 查询当前生效的 License（is_active=1，取 expires_at 最新一条）。
func (s *Service) loadActiveLicense() (*models.License, error) {
	var licenseObj models.License
	err := s.db.
		Where("is_active = ?", 1).
		Order("expires_at DESC").
		Limit(1).
		First(&licenseObj).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &licenseObj, nil
}
// countCurrentUsers 统计当前用户数。
//
// 与 Python 版对齐：统计全部用户（含禁用），与产品对"用户数"的定义一致。
func (s *Service) countCurrentUsers() (int64, error) {
	var count int64
	if err := s.db.Model(&models.User{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
// isUniqueConstraintError 判断错误是否为唯一约束冲突。
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique constraint") ||
		strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, ActiveLicenseUniqueIndex)
}
// containsString 判断切片中是否包含目标字符串。
func containsString(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
}
// validateMaxUsers 校验 max_users 字段。
//
// 与 Python 版对齐：
//   - nil 或字段缺失 → 返回 (nil, nil)；
//   - bool 类型显式拒绝；
//   - 正整数 → 返回指针；
//   - 其他 → 50001。
func validateMaxUsers(payload map[string]interface{}) (*int, error) {
	v, ok := payload["max_users"]
	if !ok || v == nil {
		return nil, nil
	}
	switch val := v.(type) {
	case bool:
		return nil, exception.New(
			exception.CodeLicenseInvalidFile,
			"License 字段 max_users 必须为正整数或 null",
			400, nil,
		)
	case float64:
		intVal := int(val)
		if float64(intVal) != val || intVal <= 0 {
			return nil, exception.New(
				exception.CodeLicenseInvalidFile,
				"License 字段 max_users 必须为正整数或 null",
				400, nil,
			)
		}
		return &intVal, nil
	case int:
		if val <= 0 {
			return nil, exception.New(
				exception.CodeLicenseInvalidFile,
				"License 字段 max_users 必须为正整数或 null",
				400, nil,
			)
		}
		return &val, nil
	case int64:
		if val <= 0 {
			return nil, exception.New(
				exception.CodeLicenseInvalidFile,
				"License 字段 max_users 必须为正整数或 null",
				400, nil,
			)
		}
		intVal := int(val)
		return &intVal, nil
	default:
		return nil, exception.New(
			exception.CodeLicenseInvalidFile,
			"License 字段 max_users 必须为正整数或 null",
			400, nil,
		)
	}
}
// validateAuthorizedModules 校验 authorized_modules 字段。
func validateAuthorizedModules(payload map[string]interface{}) ([]string, error) {
	v, ok := payload["authorized_modules"]
	if !ok || v == nil {
		return []string{}, nil
	}
	arr, ok := v.([]interface{})
	if !ok {
		return nil, exception.New(
			exception.CodeLicenseInvalidFile,
			"License 字段 authorized_modules 必须为数组",
			400, nil,
		)
	}
	result := make([]string, 0, len(arr))
	for _, item := range arr {
		s, ok := item.(string)
		if !ok || s == "" {
			return nil, exception.New(
				exception.CodeLicenseInvalidFile,
				"License 字段 authorized_modules 元素必须为非空字符串",
				400, nil,
			)
		}
		result = append(result, s)
	}
	return result, nil
}
// validateMachineCodeField 校验 machine_code 字段类型。
//
// 返回空串表示未绑定机器码。
func validateMachineCodeField(payload map[string]interface{}) string {
	v, ok := payload["machine_code"]
	if !ok || v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}
// serializeLicenseForImport 序列化 License 用于导入响应。
//
// v1.1（Go-P1-01）：AuthorizedModules 使用 normalizeAuthorizedModules 归一化，
// 保证 DB 存储 NULL 时响应为 [] 而非 null。
func serializeLicenseForImport(licenseObj *models.License) *LicenseImportResp {
	return &LicenseImportResp{
		ID:                licenseObj.ID,
		LicenseKey:        licenseObj.LicenseKey,
		LicenseType:       licenseObj.LicenseType,
		ExpiresAt:         formatDateTime(licenseObj.ExpiresAt),
		AuthorizedModules: normalizeAuthorizedModules(licenseObj.AuthorizedModules),
		MaxUsers:          licenseObj.MaxUsers,
		MachineCode:       licenseObj.MachineCode,
	}
}
// ---------------------------------------------------------------------------
// 状态查询
// ---------------------------------------------------------------------------
// GetLicenseStatus 获取当前 License 状态。
//
// 若未导入或已失效（is_active=0），返回 50009。
// 过期 License 仍返回状态，但 is_expired=true / is_valid=false。
//
// v1.1（Go-P1-01）：AuthorizedModules 使用 normalizeAuthorizedModules 归一化。
// v1.1（Go-P3-03）：使用 ptr[T] 泛型辅助函数简化指针构造。
func (s *Service) GetLicenseStatus() (*LicenseStatusOut, error) {
	licenseObj, err := s.loadActiveLicense()
	if err != nil {
		return nil, exception.New(exception.CodeInternalError, "查询 License 失败", 500, nil)
	}
	if licenseObj == nil {
		return nil, exception.New(exception.CodeLicenseNotFound, "License不存在", 404, nil)
	}
	currentMachineCode := GetMachineCode()
	currentUsers, err := s.countCurrentUsers()
	if err != nil {
		return nil, exception.New(exception.CodeInternalError, "统计用户数失败", 500, nil)
	}
	now := utcNowNaive()
	delta := licenseObj.ExpiresAt.Sub(now)
	daysRemaining := int(delta.Hours() / 24)
	if daysRemaining < 0 {
		daysRemaining = 0
	}
	isExpired := !licenseObj.ExpiresAt.After(now)
	isExpiringSoon := !isExpired && daysRemaining <= ExpiringSoonDays
	isActive := licenseObj.IsActive == 1
	machineMatch := true
	if licenseObj.MachineCode != nil && *licenseObj.MachineCode != "" {
		machineMatch = *licenseObj.MachineCode == currentMachineCode
	}
	isValid := isActive && !isExpired && machineMatch
	// 机器码仅写入服务端日志，便于排障（与 Python 版一致）
	log.Debug().
		Str("license_key", licenseObj.LicenseKey).
		Interface("bound_machine", licenseObj.MachineCode).
		Str("current_machine", currentMachineCode).
		Msg("License 状态查询")
	return &LicenseStatusOut{
		IsValid:           isValid,
		LicenseType:       ptr(licenseObj.LicenseType),
		ExpiresAt:         ptr(formatDateTime(licenseObj.ExpiresAt)),
		DaysRemaining:     ptr(daysRemaining),
		MaxUsers:          licenseObj.MaxUsers,
		CurrentUsers:      int(currentUsers),
		AuthorizedModules: normalizeAuthorizedModules(licenseObj.AuthorizedModules),
		IsExpired:         isExpired,
		IsExpiringSoon:    isExpiringSoon,
	}, nil
}
// ---------------------------------------------------------------------------
// License 落库（核心）
// ---------------------------------------------------------------------------
// persistLicense 将已校验的 payload 落库。
//
// 事务顺序（与 Python v1.2 一致）：
//  1. 字段校验；
//  2. 查询 existing；
//  3. existing.is_active==1 → 50004；
//  4. 先 UPDATE 其他 active 为 0；
//  5. 插入新记录 或 更新 existing 为 active；
//  6. commit，捕获唯一约束冲突 → 50004。
//
// 并发保证：部分唯一索引 uq_license_license_active 保证同一时刻最多一条 active。
func (s *Service) persistLicense(
	payload map[string]interface{},
	operatorID uint,
	operatorName string,
) (*models.License, error) {
	// 1. license_type 白名单
	licenseType, _ := payload["license_type"].(string)
	if !containsString(LicenseTypes, licenseType) {
		return nil, exception.New(
			exception.CodeLicenseTypeUnsupported,
			fmt.Sprintf("License 类型不支持：%s", licenseType),
			400, nil,
		)
	}
	// 2. 必需字段
	for _, field := range RequiredPayloadFields {
		v, ok := payload[field]
		if !ok || v == nil || v == "" {
			return nil, exception.New(
				exception.CodeLicenseInvalidFile,
				fmt.Sprintf("License 缺少必需字段：%s", field),
				400, nil,
			)
		}
	}
	// 3. license_key 长度
	licenseKey, _ := payload["license_key"].(string)
	if len(licenseKey) < LicenseKeyMinLen || len(licenseKey) > LicenseKeyMaxLen {
		return nil, exception.New(
			exception.CodeLicenseInvalidFile,
			fmt.Sprintf("License 字段 license_key 长度必须为 %d-%d 位", LicenseKeyMinLen, LicenseKeyMaxLen),
			400, nil,
		)
	}
	// 4. 时间解析与逻辑校验
	issuedAt, err := parseISODateTime(payload["issued_at"], "issued_at")
	if err != nil {
		return nil, err
	}
	expiresAt, err := parseISODateTime(payload["expires_at"], "expires_at")
	if err != nil {
		return nil, err
	}
	if !issuedAt.Before(expiresAt) {
		return nil, exception.New(
			exception.CodeLicenseInvalidFile,
			"License 字段 issued_at 必须早于 expires_at",
			400, nil,
		)
	}
	now := utcNowNaive()
	if !expiresAt.After(now) {
		return nil, exception.New(exception.CodeLicenseExpired, "License 已过期", 400, nil)
	}
	// 5. machine_code 校验
	currentMachineCode := GetMachineCode()
	boundMachine := validateMachineCodeField(payload)
	if boundMachine != "" && boundMachine != currentMachineCode {
		return nil, exception.New(
			exception.CodeLicenseMachineMismatch,
			fmt.Sprintf("License 绑定机器码不匹配（license=%s, current=%s）", boundMachine, currentMachineCode),
			400, nil,
		)
	}
	// 6. max_users 校验
	maxUsers, err := validateMaxUsers(payload)
	if err != nil {
		return nil, err
	}
	// 7. authorized_modules 校验
	authorizedModules, err := validateAuthorizedModules(payload)
	if err != nil {
		return nil, err
	}
	// 8. 查询 existing
	var existing models.License
	err = s.db.Where("license_key = ?", licenseKey).First(&existing).Error
	existingExists := err == nil
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.New(exception.CodeInternalError, "查询 License 失败", 500, nil)
	}
	// 9. existing 且 is_active==1 → 50004
	if existingExists && existing.IsActive == 1 {
		return nil, exception.New(exception.CodeLicenseAlreadyActive, "License 已激活", 409, nil)
	}
	// 10-11. 事务：先 UPDATE 其他 active，再插入/更新当前
	var result *models.License
	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		// 10. 先把其他 active 置 0
		if err := tx.Model(&models.License{}).
			Where("is_active = ? AND license_key != ?", 1, licenseKey).
			Updates(map[string]interface{}{
				"is_active":  0,
				"updated_at": now,
			}).Error; err != nil {
			return err
		}
		// 11. 插入或更新
		if existingExists {
			existing.LicenseType = licenseType
			existing.MaxUsers = maxUsers
			existing.AuthorizedModules = models.StringArray(authorizedModules)
			if boundMachine != "" {
				machineCopy := boundMachine
				existing.MachineCode = &machineCopy
			} else {
				existing.MachineCode = nil
			}
			existing.IssuedAt = issuedAt
			existing.ExpiresAt = expiresAt
			existing.IsActive = 1
			existing.ActivatedAt = &now
			existing.UpdatedAt = now
			if err := tx.Save(&existing).Error; err != nil {
				return err
			}
			result = &existing
			return nil
		}
		newLicense := models.License{
			LicenseKey:        licenseKey,
			LicenseType:       licenseType,
			MaxUsers:          maxUsers,
			AuthorizedModules: models.StringArray(authorizedModules),
			IssuedAt:          issuedAt,
			ExpiresAt:         expiresAt,
			IsActive:          1,
			ActivatedAt:       &now,
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		if boundMachine != "" {
			machineCopy := boundMachine
			newLicense.MachineCode = &machineCopy
		}
		if err := tx.Create(&newLicense).Error; err != nil {
			return err
		}
		result = &newLicense
		return nil
	})
	if txErr != nil {
		if isUniqueConstraintError(txErr) {
			return nil, exception.New(exception.CodeLicenseAlreadyActive, "License 已激活", 409, nil)
		}
		var pe *exception.PlatformError
		if errors.As(txErr, &pe) {
			return nil, pe
		}
		return nil, exception.New(
			exception.CodeInternalError,
			fmt.Sprintf("License 保存失败：%v", txErr),
			500, nil,
		)
	}
	// 审计日志
	if operatorID != 0 {
		auth.LogAuthEvent(
			"license_import",
			&operatorID, operatorName, "success", nil, "", "",
			fmt.Sprintf(
				"导入 License %s（type=%s, modules=%d, expires=%s）",
				licenseKey, licenseType, len(authorizedModules),
				expiresAt.Format(time.RFC3339),
			),
		)
	}
	log.Info().
		Str("license_key", licenseKey).
		Str("license_type", licenseType).
		Time("expires_at", expiresAt).
		Msg("License 导入成功")
	return result, nil
}
// ---------------------------------------------------------------------------
// 导入 / 激活
// ---------------------------------------------------------------------------
// ImportLicense 导入 License。
//
// 两种方式（二选一）：
//   - fileContent：License 文件内容（.lic / .json）；
//   - activationCode：在线激活码。
func (s *Service) ImportLicense(
	fileContent []byte,
	activationCode string,
	operatorID uint,
	operatorName string,
) (*models.License, error) {
	hasFile := len(fileContent) > 0
	hasCode := strings.TrimSpace(activationCode) != ""
	if !hasFile && !hasCode {
		return nil, exception.New(
			exception.CodeLicenseInvalidFile,
			"必须提供 license_file 或 activation_code",
			400, nil,
		)
	}
	var content []byte
	if hasFile {
		content = fileContent
	} else {
		fetched, err := s.fetchLicenseByActivationCode(activationCode, GetMachineCode())
		if err != nil {
			return nil, err
		}
		content = fetched
	}
	payload, signature, err := ParseLicenseContent(content)
	if err != nil {
		return nil, err
	}
	valid, err := VerifySignature(payload, signature, s.cfg.LicenseSecretKey)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError, "签名验证失败", 500, nil)
	}
	if !valid {
		return nil, exception.New(
			exception.CodeLicenseSignatureFailed,
			"License 签名验证失败",
			400, nil,
		)
	}
	return s.persistLicense(payload, operatorID, operatorName)
}
// ActivateLicense 在线激活。
//
// 前置校验：
//   - 请求 machine_code 必须与当前机器码一致，否则 → 50006；
//   - 强制使用当前机器码请求外部服务（防篡改）。
func (s *Service) ActivateLicense(
	activationCode, machineCode string,
	operatorID uint,
	operatorName string,
) (*models.License, error) {
	currentMachineCode := GetMachineCode()
	if machineCode != currentMachineCode {
		return nil, exception.New(
			exception.CodeLicenseMachineMismatch,
			fmt.Sprintf("机器码不匹配（请求=%s, 当前=%s）", machineCode, currentMachineCode),
			400, nil,
		)
	}
	content, err := s.fetchLicenseByActivationCode(activationCode, currentMachineCode)
	if err != nil {
		return nil, err
	}
	payload, signature, err := ParseLicenseContent(content)
	if err != nil {
		return nil, err
	}
	valid, err := VerifySignature(payload, signature, s.cfg.LicenseSecretKey)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError, "签名验证失败", 500, nil)
	}
	if !valid {
		return nil, exception.New(
			exception.CodeLicenseSignatureFailed,
			"License 签名验证失败",
			400, nil,
		)
	}
	return s.persistLicense(payload, operatorID, operatorName)
}
// fetchLicenseByActivationCode 调用外部激活服务获取 License 文件内容。
func (s *Service) fetchLicenseByActivationCode(
	activationCode, machineCode string,
) ([]byte, error) {
	url := strings.TrimSpace(s.cfg.LicenseActivationURL)
	if url == "" {
		return nil, exception.New(
			exception.CodeLicenseActivationCode,
			"激活码无效（未配置在线激活服务）",
			400, nil,
		)
	}
	requestBody := map[string]string{
		"activation_code": activationCode,
		"machine_code":    machineCode,
	}
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, exception.New(
			exception.CodeLicenseActivationCode,
			fmt.Sprintf("构造激活请求失败：%v", err),
			400, nil,
		)
	}
	resp, err := s.httpClient.Post(
		strings.TrimRight(url, "/")+"/activate",
		"application/json",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return nil, exception.New(
			exception.CodeLicenseActivationCode,
			fmt.Sprintf("在线激活服务不可用：%v", err),
			400, nil,
		)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, exception.New(
			exception.CodeLicenseActivationCode,
			fmt.Sprintf("在线激活失败（HTTP %d）", resp.StatusCode),
			400, nil,
		)
	}
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, exception.New(
			exception.CodeLicenseActivationCode,
			fmt.Sprintf("读取激活响应失败：%v", err),
			400, nil,
		)
	}
	var body map[string]interface{}
	if err := json.Unmarshal(respBytes, &body); err != nil {
		return nil, exception.New(
			exception.CodeLicenseActivationCode,
			fmt.Sprintf("在线激活响应不是合法 JSON：%v", err),
			400, nil,
		)
	}
	code, _ := body["code"].(float64)
	if code != 0 {
		msg, _ := body["message"].(string)
		if msg == "" {
			msg = "未知错误"
		}
		return nil, exception.New(
			exception.CodeLicenseActivationCode,
			fmt.Sprintf("激活码无效：%s", msg),
			400, nil,
		)
	}
	data, ok := body["data"].(map[string]interface{})
	if !ok {
		return nil, exception.New(
			exception.CodeLicenseActivationCode,
			"在线激活响应缺少 data",
			400, nil,
		)
	}
	if _, ok := data["payload"].(map[string]interface{}); !ok {
		return nil, exception.New(
			exception.CodeLicenseActivationCode,
			"在线激活响应缺少 payload 或 signature",
			400, nil,
		)
	}
	if _, ok := data["signature"].(string); !ok {
		return nil, exception.New(
			exception.CodeLicenseActivationCode,
			"在线激活响应缺少 payload 或 signature",
			400, nil,
		)
	}
	return json.Marshal(data)
}
// ---------------------------------------------------------------------------
// 模块授权
// ---------------------------------------------------------------------------
// GetModuleAuthorization 获取所有已安装模块的授权状态。
func (s *Service) GetModuleAuthorization() ([]*ModuleAuthorizationOut, error) {
	licenseObj, err := s.loadActiveLicense()
	if err != nil {
		return nil, exception.New(exception.CodeInternalError, "查询 License 失败", 500, nil)
	}
	authorizedModules := map[string]struct{}{}
	var expiresAt *time.Time
	if licenseObj != nil {
		// for range nil slice 在 Go 中安全（不会 panic），此处无需归一化
		for _, m := range licenseObj.AuthorizedModules {
			authorizedModules[m] = struct{}{}
		}
		expiresAt = &licenseObj.ExpiresAt
	}
	var modules []models.Module
	if err := s.db.Order("id ASC").Find(&modules).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "查询模块失败", 500, nil)
	}
	result := make([]*ModuleAuthorizationOut, 0, len(modules))
	for _, m := range modules {
		_, isAuth := authorizedModules[m.ID]
		_, isCore := CoreModulesBypass[m.ID]
		isAuthorized := isAuth || isCore
		var expiresAtStr *string
		if expiresAt != nil {
			s := formatDateTime(*expiresAt)
			expiresAtStr = &s
		}
		result = append(result, &ModuleAuthorizationOut{
			ModuleID:     m.ID,
			ModuleName:   m.Name,
			IsAuthorized: isAuthorized,
			ExpiresAt:    expiresAtStr,
		})
	}
	return result, nil
}
// CheckModuleAuthorized 校验模块是否已授权。
//
// 核心模块直接放行；业务模块需查 active license 是否包含该 module_id。
func (s *Service) CheckModuleAuthorized(moduleID string) error {
	if _, ok := CoreModulesBypass[moduleID]; ok {
		return nil
	}
	licenseObj, err := s.loadActiveLicense()
	if err != nil {
		return exception.New(exception.CodeInternalError, "查询 License 失败", 500, nil)
	}
	if licenseObj == nil {
		return exception.New(exception.CodeLicenseNotFound, "License不存在", 404, nil)
	}
	now := utcNowNaive()
	if !licenseObj.ExpiresAt.After(now) {
		return exception.New(exception.CodeLicenseExpired, "License 已过期", 400, nil)
	}
	if licenseObj.MachineCode != nil && *licenseObj.MachineCode != "" &&
		*licenseObj.MachineCode != GetMachineCode() {
		return exception.New(exception.CodeLicenseMachineMismatch, "机器码不匹配", 400, nil)
	}
	for _, m := range licenseObj.AuthorizedModules {
		if m == moduleID {
			return nil
		}
	}
	return exception.New(
		exception.CodeLicenseModuleNotAuth,
		fmt.Sprintf("模块未授权：%s", moduleID),
		403, nil,
	)
}
// ---------------------------------------------------------------------------
// 用户配额
// ---------------------------------------------------------------------------
// CheckUserQuota 校验当前用户数是否超过 License 限制。
//
// 与 Python 版对齐：统计全部用户（含禁用）。
func (s *Service) CheckUserQuota() error {
	licenseObj, err := s.loadActiveLicense()
	if err != nil {
		return exception.New(exception.CodeInternalError, "查询 License 失败", 500, nil)
	}
	if licenseObj == nil {
		return exception.New(exception.CodeLicenseNotFound, "License不存在", 404, nil)
	}
	now := utcNowNaive()
	if !licenseObj.ExpiresAt.After(now) {
		return exception.New(exception.CodeLicenseExpired, "License 已过期", 400, nil)
	}
	if licenseObj.MaxUsers == nil {
		return nil
	}
	currentUsers, err := s.countCurrentUsers()
	if err != nil {
		return exception.New(exception.CodeInternalError, "统计用户数失败", 500, nil)
	}
	if int(currentUsers) >= *licenseObj.MaxUsers {
		return exception.New(
			exception.CodeLicenseUserQuotaExceed,
			fmt.Sprintf("用户数超限：当前 %d，License 上限 %d", currentUsers, *licenseObj.MaxUsers),
			400, nil,
		)
	}
	return nil
}
// ---------------------------------------------------------------------------
// 启动校验
// ---------------------------------------------------------------------------
// VerifyLicenseOnStartup 启动时校验 License 状态。
//
// 不阻断启动（允许运维通过 API 导入/更新 License）。
// 返回状态字典，供调用方记录日志或决定是否阻断。
func (s *Service) VerifyLicenseOnStartup() map[string]interface{} {
	status, err := s.GetLicenseStatus()
	if err != nil {
		var pe *exception.PlatformError
		if errors.As(err, &pe) && pe.Code == exception.CodeLicenseNotFound {
			log.Warn().Msg("启动校验：未导入 License（code=50009）")
			return map[string]interface{}{
				"ok": false, "code": pe.Code, "message": pe.Message,
			}
		}
		log.Error().Err(err).Msg("启动校验失败")
		return map[string]interface{}{
			"ok": false, "code": 90000, "message": err.Error(),
		}
	}
	if status.IsExpired {
		log.Error().
			Interface("expires_at", status.ExpiresAt).
			Msg("启动校验：License 已过期")
	} else if !status.IsValid {
		log.Error().Msg("启动校验：License 无效（machine_code 不匹配或 is_active=0）")
	} else if status.IsExpiringSoon {
		log.Warn().
			Interface("days_remaining", status.DaysRemaining).
			Msg("启动校验：License 即将过期")
	} else {
		log.Info().
			Interface("license_type", status.LicenseType).
			Interface("days_remaining", status.DaysRemaining).
			Msg("启动校验：License 有效")
	}
	return map[string]interface{}{"ok": true, "status": status}
}
// SerializeLicenseForImport 对外暴露序列化辅助（供 handler 使用）。
func SerializeLicenseForImport(licenseObj *models.License) *LicenseImportResp {
	return serializeLicenseForImport(licenseObj)
}
