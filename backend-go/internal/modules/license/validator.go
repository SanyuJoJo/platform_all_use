package license
import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"runtime"
	"strings"
	"backend-go/internal/config"
	"backend-go/internal/exception"
)
// CanonicalJSON 复刻 Python 的 canonical_json 序列化。
//
// 等价于：json.dumps(payload, sort_keys=True, separators=(",", ":"), ensure_ascii=False)
//
// 关键点：
//   - Go 的 encoding/json 对 map 默认按键字母序序列化（自 Go 1.12 起稳定）；
//   - SetEscapeHTML(false) 避免 < > & 被转义为 \u003c 等；
//   - 去掉 Encode 追加的尾部换行符，得到紧凑格式。
func CanonicalJSON(payload map[string]interface{}) (string, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(payload); err != nil {
		return "", err
	}
	return strings.TrimRight(buf.String(), "\n"), nil
}
// ComputeSignature 计算 payload 的 HMAC-SHA256 签名。
func ComputeSignature(payload map[string]interface{}, secret string) (string, error) {
	canonical, err := CanonicalJSON(payload)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(canonical))
	return hex.EncodeToString(mac.Sum(nil)), nil
}
// VerifySignature 使用 hmac.Equal 防止时序攻击。
func VerifySignature(payload map[string]interface{}, signature, secret string) (bool, error) {
	expected, err := ComputeSignature(payload, secret)
	if err != nil {
		return false, err
	}
	return hmac.Equal([]byte(expected), []byte(signature)), nil
}
// GetMachineCode 生成当前机器码。
//
// 优先级：
//  1. config.C.LicenseMachineCodeOverride（由 config.Load() 从环境变量
//     LICENSE_MACHINE_CODE_OVERRIDE 读取）；
//  2. 基于 MAC 地址 + 平台信息生成，SHA-256 后取前 32 位大写 hex。
//
// v1.1（Go-P3-01）：删除 os.Getenv 兜底逻辑，统一通过 config.C 读取。
//   - config.Load() 中已通过 v.AutomaticEnv() 加载所有环境变量；
//   - 双路径读取会造成维护负担且可能因读取时机不同产生不一致。
//
// 跨语言一致性说明：
//   - Python 的 platform.system() / platform.machine() 返回的值与 Go 的
//     runtime.GOOS / runtime.GOARCH 命名不完全一致，本实现在主流平台上
//     做了映射（见 platformSystem / platformMachine）；
//   - 若在非主流发行版或容器环境中出现机器码不一致，可通过
//     LICENSE_MACHINE_CODE_OVERRIDE 固定。
func GetMachineCode() string {
	override := ""
	if config.C != nil {
		override = strings.TrimSpace(config.C.LicenseMachineCodeOverride)
	}
	if override != "" {
		return override
	}
	mac := getMACAddress()
	parts := []string{
		fmt.Sprintf("%012x", mac),
		platformSystem(),
		platformMachine(),
	}
	raw := strings.Join(parts, "|")
	hash := sha256.Sum256([]byte(raw))
	// SHA-256 hex 输出恒为 64 字符，直接切片取前 32 位（Go-P3-04）
	return strings.ToUpper(hex.EncodeToString(hash[:])[:32])
}
// getMACAddress 获取首块非 loopback 网卡的 MAC 地址。
//
// 与 Python 的 uuid.getnode() 语义接近，优先返回物理网卡 MAC。
func getMACAddress() uint64 {
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range interfaces {
			if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
				continue
			}
			if len(iface.HardwareAddr) >= 6 {
				var mac uint64
				for _, b := range iface.HardwareAddr[:6] {
					mac = mac<<8 | uint64(b)
				}
				return mac
			}
		}
	}
	return 0
}
// platformSystem 与 Python platform.system() 命名对齐。
func platformSystem() string {
	switch runtime.GOOS {
	case "linux":
		return "Linux"
	case "darwin":
		return "Darwin"
	case "windows":
		return "Windows"
	default:
		return runtime.GOOS
	}
}
// platformMachine 与 Python platform.machine() 命名对齐。
//
// 注意：Linux 下 Python 返回 "aarch64"，macOS 下 Python 返回 "arm64"，
// 但 Go 的 runtime.GOARCH 在两种系统上都为 "arm64"，故需区分 GOOS。
func platformMachine() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x86_64"
	case "arm64":
		if runtime.GOOS == "darwin" {
			return "arm64"
		}
		return "aarch64"
	case "386":
		return "i386"
	default:
		return runtime.GOARCH
	}
}
// ParseLicenseContent 解析 License 文件内容，返回 (payload, signature, error)。
//
// 失败时抛出 50001。
func ParseLicenseContent(content []byte) (map[string]interface{}, string, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, "", exception.New(
			exception.CodeLicenseInvalidFile,
			fmt.Sprintf("License 文件不是合法 JSON：%v", err),
			400, nil,
		)
	}
	payload, ok := data["payload"].(map[string]interface{})
	if !ok {
		return nil, "", exception.New(
			exception.CodeLicenseInvalidFile,
			"License 文件缺少 payload", 400, nil,
		)
	}
	signature, ok := data["signature"].(string)
	if !ok || signature == "" {
		return nil, "", exception.New(
			exception.CodeLicenseInvalidFile,
			"License 文件缺少 signature", 400, nil,
		)
	}
	return payload, signature, nil
}
