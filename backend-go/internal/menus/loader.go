package menus

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"

	"backend-go/internal/exception"
)

var permissionCodePattern = regexp.MustCompile(`^[a-z0-9_]+:[a-z0-9_]+:[a-z0-9_]+$`)

// Load 从文件加载并校验前端配置。
func Load(path string) (*FrontendConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, exception.New(
			exception.CodeInternalError,
			fmt.Sprintf("读取前端配置失败：%v", err), 500, nil,
		)
	}

	cfg := &FrontendConfig{}
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, exception.New(
				exception.CodeInternalError,
				fmt.Sprintf("解析 YAML 失败：%v", err), 500, nil,
			)
		}
	case ".json":
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, exception.New(
				exception.CodeInternalError,
				fmt.Sprintf("解析 JSON 失败：%v", err), 500, nil,
			)
		}
	default:
		return nil, exception.New(
			exception.CodeInternalError,
			fmt.Sprintf("不支持的文件扩展名：%s", ext), 500, nil,
		)
	}

	if err := Validate(cfg); err != nil {
		return nil, err
	}

	totalMenus := 0
	for _, m := range cfg.ModuleMenus {
		totalMenus += countMenuItems(m)
	}

	log.Info().
		Str("path", path).
		Int("modules_with_menus", len(cfg.ModuleMenus)).
		Int("total_menus", totalMenus).
		Int("menus_hidden", len(cfg.MenusHidden)).
		Int("entry_overrides", len(cfg.EntryFrontendOverrides)).
		Msg("前端配置加载完成")
	return cfg, nil
}

// Validate 校验配置。
//
// 校验项：
//  1. module_menus 中每个模块的菜单：
//     - 该模块内部 id 唯一；
//     - title / path 非空；
//     - permission（若非空）符合三段式小写；
//  2. menus_hidden：模块 ID 非空；不在 module_menus 中重复（冲突时以 hidden 为准，警告）；
//  3. entry_frontend_overrides：key 与 value 均非空。
func Validate(cfg *FrontendConfig) error {
	if cfg == nil {
		return exception.New(exception.CodeInternalError, "前端配置为空", 400, nil)
	}

	// 1. module_menus
	for moduleID, menus := range cfg.ModuleMenus {
		if strings.TrimSpace(moduleID) == "" {
			return fieldErr("module_menus", "存在空模块 ID")
		}
		menuIDs := make(map[string]struct{})
		prefix := fmt.Sprintf("module_menus.%s", moduleID)
		if err := walkMenus(menus, prefix, menuIDs); err != nil {
			return err
		}
	}

	// 2. menus_hidden
	hiddenSet := make(map[string]struct{}, len(cfg.MenusHidden))
	for i, id := range cfg.MenusHidden {
		if strings.TrimSpace(id) == "" {
			return fieldErr(fmt.Sprintf("menus_hidden[%d]", i), "不能为空")
		}
		if _, dup := hiddenSet[id]; dup {
			return fieldErr(fmt.Sprintf("menus_hidden[%d]", i), "重复："+id)
		}
		hiddenSet[id] = struct{}{}
	}
	// 与 module_menus 冲突检查（警告）
	for _, id := range cfg.MenusHidden {
		if _, conflict := cfg.ModuleMenus[id]; conflict {
			log.Warn().
				Str("module_id", id).
				Msg("模块同时出现在 menus_hidden 与 module_menus 中，以 menus_hidden 为准")
		}
	}

	// 3. entry_frontend_overrides
	for k, v := range cfg.EntryFrontendOverrides {
		if strings.TrimSpace(k) == "" {
			return fieldErr("entry_frontend_overrides", "存在空 key")
		}
		if strings.TrimSpace(v) == "" {
			return fieldErr("entry_frontend_overrides."+k, "不能为空")
		}
	}
	return nil
}

func walkMenus(items []interface{}, prefix string, menuIDs map[string]struct{}) error {
	for i, raw := range items {
		m, ok := raw.(map[string]interface{})
		if !ok {
			if m2, ok2 := raw.(map[interface{}]interface{}); ok2 {
				m = normalizeMap(m2)
			} else {
				return fieldErr(fmt.Sprintf("%s[%d]", prefix, i), "菜单项必须为对象")
			}
		}

		id := toString(m["id"])
		if id == "" {
			return fieldErr(fmt.Sprintf("%s[%d].id", prefix, i), "不能为空")
		}
		if _, dup := menuIDs[id]; dup {
			return fieldErr(fmt.Sprintf("%s[%d].id", prefix, i), "重复："+id)
		}
		menuIDs[id] = struct{}{}

		title := toString(m["title"])
		if title == "" {
			return fieldErr(fmt.Sprintf("%s[%d].title", prefix, i), "不能为空")
		}
		path := toString(m["path"])
		if path == "" {
			return fieldErr(fmt.Sprintf("%s[%d].path", prefix, i), "不能为空")
		}

		if perm := toString(m["permission"]); perm != "" && !permissionCodePattern.MatchString(perm) {
			return fieldErr(fmt.Sprintf("%s[%d].permission", prefix, i),
				"格式无效（应为 {module}:{resource}:{action}）："+perm)
		}

		if children, ok := m["children"]; ok {
			childSlice, ok := children.([]interface{})
			if !ok {
				return fieldErr(fmt.Sprintf("%s[%d].children", prefix, i), "必须为数组")
			}
			if err := walkMenus(childSlice,
				fmt.Sprintf("%s[%d].children", prefix, i), menuIDs); err != nil {
				return err
			}
		}
	}
	return nil
}

// countMenuItems 递归统计菜单项数量。
func countMenuItems(items []interface{}) int {
	n := len(items)
	for _, raw := range items {
		m, ok := raw.(map[string]interface{})
		if !ok {
			if m2, ok2 := raw.(map[interface{}]interface{}); ok2 {
				m = normalizeMap(m2)
			} else {
				continue
			}
		}
		if children, ok := m["children"].([]interface{}); ok {
			n += countMenuItems(children)
		}
	}
	return n
}

// normalizeMap 将 map[interface{}]interface{} 转为 map[string]interface{}。
func normalizeMap(in map[interface{}]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		key := fmt.Sprintf("%v", k)
		switch val := v.(type) {
		case map[interface{}]interface{}:
			out[key] = normalizeMap(val)
		case []interface{}:
			out[key] = normalizeSlice(val)
		default:
			out[key] = v
		}
	}
	return out
}

func normalizeSlice(in []interface{}) []interface{} {
	out := make([]interface{}, len(in))
	for i, v := range in {
		switch val := v.(type) {
		case map[interface{}]interface{}:
			out[i] = normalizeMap(val)
		case []interface{}:
			out[i] = normalizeSlice(val)
		default:
			out[i] = v
		}
	}
	return out
}

func fieldErr(field, msg string) error {
	return exception.New(
		exception.CodeInternalError,
		fmt.Sprintf("前端配置校验失败：%s %s", field, msg), 400, nil,
	)
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
