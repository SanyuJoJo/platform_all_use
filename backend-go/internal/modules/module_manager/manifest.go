package module_manager
import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"backend-go/internal/exception"
)
var (
	moduleIDPattern       = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	permissionCodePattern = regexp.MustCompile(`^[a-z0-9_]+:[a-z0-9_]+:[a-z0-9_]+$`)
	menuIDPattern         = regexp.MustCompile(`^[a-z][a-z0-9_]*:[a-z][a-z0-9_]*$`)
	semverPattern         = regexp.MustCompile(`^\d+\.\d+\.\d+(-[a-zA-Z0-9.]+)?(\+[a-zA-Z0-9.]+)?$`)
	tableNamePattern      = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
)
// Manifest 模块清单结构。
type Manifest struct {
	ID             string                   `json:"id"`
	Name           string                   `json:"name"`
	Version        string                   `json:"version"`
	Description    string                   `json:"description"`
	Author         *string                  `json:"author"`
	Homepage       *string                  `json:"homepage"`
	Dependencies   []string                 `json:"dependencies"`
	Permissions    []map[string]interface{} `json:"permissions"`
	Menus          []map[string]interface{} `json:"menus"`
	ConfigSchema   map[string]interface{}   `json:"config_schema"`
	EntryBackend   string                   `json:"entry_backend"`
	EntryFrontend  *string                  `json:"entry_frontend"`
	DatabaseTables []string                 `json:"database_tables"`
}
// LoadManifestFromDir 从目录读取并校验 manifest.json，返回原始 map。
func LoadManifestFromDir(moduleDir string) (map[string]interface{}, error) {
	manifestPath := moduleDir + "/manifest.json"
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, exception.New(exception.CodeModuleManifestInvalid,
			"模块清单文件不存在", 400, nil)
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, exception.New(exception.CodeModuleManifestInvalid,
			fmt.Sprintf("模块清单文件格式错误：%v", err), 400, nil)
	}
	if _, err := ValidateManifestDict(raw); err != nil {
		return nil, err
	}
	return raw, nil
}
// ValidateManifestDict 校验 manifest 字典，返回 Manifest 对象。
func ValidateManifestDict(raw map[string]interface{}) (*Manifest, error) {
	bytes, err := json.Marshal(raw)
	if err != nil {
		return nil, exception.New(exception.CodeModuleManifestInvalid,
			fmt.Sprintf("模块清单序列化失败：%v", err), 400, nil)
	}
	var m Manifest
	if err := json.Unmarshal(bytes, &m); err != nil {
		return nil, exception.New(exception.CodeModuleManifestInvalid,
			fmt.Sprintf("模块清单校验失败：%v", err), 400, nil)
	}
	if err := validateManifest(&m); err != nil {
		return nil, err
	}
	return &m, nil
}
// validateManifest 业务级校验。
//
// v1.1（P0-12）：移除文件存在性校验（Go 语境下不成立），
// 仅校验 entry_backend 格式为 `{package_path}:{symbol_name}`。
func validateManifest(m *Manifest) error {
	// 1. 模块 ID 格式
	if !moduleIDPattern.MatchString(m.ID) {
		return exception.New(exception.CodeModuleIDInvalid,
			"模块 ID 格式无效（小写字母开头，仅允许小写字母、数字、下划线）", 400, nil)
	}
	// 2. 版本号
	if !semverPattern.MatchString(m.Version) {
		return exception.New(exception.CodeModuleManifestInvalid,
			fmt.Sprintf("模块版本号格式无效（应为 x.y.z 或 x.y.z-pre/build）：%s", m.Version), 400, nil)
	}
	// 3. 依赖 ID 格式
	for _, dep := range m.Dependencies {
		if !moduleIDPattern.MatchString(dep) {
			return exception.New(exception.CodeModuleManifestInvalid,
				fmt.Sprintf("依赖模块 ID 格式无效：%s", dep), 400, nil)
		}
	}
	// 4. 权限编码
	for _, perm := range m.Permissions {
		code, _ := perm["code"].(string)
		if !permissionCodePattern.MatchString(code) {
			return exception.New(exception.CodeModuleManifestInvalid,
				fmt.Sprintf("权限编码格式无效：%s", code), 400, nil)
		}
		parts := strings.Split(code, ":")
		if len(parts) != 3 || parts[0] != m.ID {
			return exception.New(exception.CodeModuleManifestInvalid,
				fmt.Sprintf("权限编码前缀必须与模块 ID 一致：%s", code), 400, nil)
		}
	}
	// 5. 菜单 ID
	for _, menu := range m.Menus {
		menuID, _ := menu["id"].(string)
		if !menuIDPattern.MatchString(menuID) {
			return exception.New(exception.CodeModuleManifestInvalid,
				fmt.Sprintf("菜单 ID 格式无效（应为 {module_id}:{menu_id}）：%s", menuID), 400, nil)
		}
		if !strings.HasPrefix(menuID, m.ID+":") {
			return exception.New(exception.CodeModuleManifestInvalid,
				fmt.Sprintf("菜单 ID 前缀必须为 %s：%s", m.ID, menuID), 400, nil)
		}
	}
	// 6. 数据库表名
	for _, table := range m.DatabaseTables {
		if !tableNamePattern.MatchString(table) {
			return exception.New(exception.CodeModuleManifestInvalid,
				fmt.Sprintf("数据库表名格式无效：%s", table), 400, nil)
		}
		if !strings.HasPrefix(table, m.ID+"_") {
			return exception.New(exception.CodeModuleManifestInvalid,
				fmt.Sprintf("数据库表名必须以 %s_ 开头：%s", m.ID, table), 400, nil)
		}
	}
	// 7. entry_backend 格式校验
	//
	// v1.1（P0-12）：Go 无"入口文件"概念。
	// P3 阶段仅校验格式 `{package_path}:{symbol_name}`；
	// P6 阶段接入子进程方案后再补可达性校验。
	parts := strings.SplitN(m.EntryBackend, ":", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return exception.New(exception.CodeModuleEntryNotFound,
			"entry_backend 格式无效，应为 {package_path}:{symbol_name}", 400, nil)
	}
	return nil
}
