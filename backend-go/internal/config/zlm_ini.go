package config

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ZLMSampleSecret ZLM 官方示例 secret。
//
// 该值会被 ZLM 启动时自动检测并随机替换（写回 config.ini）。
// 后端在读取 ini 时若发现是此值，主动生成随机值写回，避免启动后被改写。
const ZLMSampleSecret = "035c73f7-bb6b-4889-a715-d9eb2d1925cc"

// ZLMINIConfig 从 config.ini 读出的关键字段。
type ZLMINIConfig struct {
	HTTPPort  int    // [http] port
	APISecret string // [api] secret
}

// ReadZLMConfigINI 从 config.ini 读取 [http] port 与 [api] secret。
//
// 解析规则（与 ZLM 内部 INI 解析器保持一致）：
//   - 空行、# / ; 开头注释行忽略
//   - [section] 切换当前段（大小写不敏感）
//   - key=value 解析，行内 # 之后视为注释
//   - 同名段多次出现时，后出现的值覆盖前面（符合 ZLM 的 INI 合并语义）
func ReadZLMConfigINI(path string) (*ZLMINIConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cfg := &ZLMINIConfig{}
	section := ""
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.ToLower(strings.Trim(line, "[] "))
			continue
		}
		idx := strings.Index(line, "=")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		if i := strings.Index(val, "#"); i >= 0 {
			val = strings.TrimSpace(val[:i])
		}
		switch section {
		case "http":
			if key == "port" {
				if n, err := strconv.Atoi(val); err == nil && n > 0 {
					cfg.HTTPPort = n
				}
			}
		case "api":
			if key == "secret" {
				cfg.APISecret = val
			}
		}
	}
	return cfg, scanner.Err()
}

// UpdateZLMConfigINI 更新 config.ini 中指定 section 的 key。
//
// 若该 key 在 section 内已存在则原地替换；
// 否则追加到该 section 的最后一个同名段末尾（找不到 section 则在文件尾追加）。
func UpdateZLMConfigINI(path, section, key, value string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")
	curSection := ""
	matched := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
			continue
		}
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			curSection = strings.ToLower(strings.Trim(trimmed, "[] "))
			continue
		}
		if curSection != section {
			continue
		}
		idx := strings.Index(line, "=")
		if idx < 0 {
			continue
		}
		k := strings.TrimSpace(line[:idx])
		if k == key {
			lines[i] = key + "=" + value
			matched = true
		}
	}
	if !matched {
		lastIdx := -1
		for i, line := range lines {
			if strings.ToLower(strings.TrimSpace(line)) == "["+section+"]" {
				lastIdx = i
			}
		}
		if lastIdx >= 0 {
			lines = append(lines[:lastIdx+1],
				append([]string{key + "=" + value}, lines[lastIdx+1:]...)...)
		} else {
			lines = append(lines, "["+section+"]", key+"="+value)
		}
	}
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644)
}

// GenerateHexSecret 生成 n 字节的 hex 随机串。
func GenerateHexSecret(nBytes int) string {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", os.Getpid())
	}
	return hex.EncodeToString(b)
}

// EnvExplicit 判断环境变量是否显式设置（非空）。
//
// 用于区分"用户显式配置"与"viper 的默认值"：
// viper.GetString 无法区分，需直接读 os.Getenv。
func EnvExplicit(key string) bool {
	return strings.TrimSpace(os.Getenv(key)) != ""
}
