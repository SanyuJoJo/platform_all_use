package manager

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// copyFile 复制文件，自动创建目标目录，并赋予 0o755 权限。
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o755)
}

// verifyChecksum 校验文件 SHA256；expected 为空则跳过校验。
func verifyChecksum(path, expected string) error {
	if expected == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	h := sha256.Sum256(data)
	actual := hex.EncodeToString(h[:])
	if actual != expected {
		return fmt.Errorf("checksum mismatch for %s: expected=%s actual=%s", path, expected, actual)
	}
	return nil
}

// resolveEntrypoint 按当前平台解析插件入口路径。
//
// 规则：
//  1. entry 为空 → 报错。
//  2. entry 为绝对路径 → 直接返回。
//  3. 尝试 dir/entry。
//  4. Windows 下若 entry 未带 .exe 后缀，再尝试 dir/entry.exe。
//  5. 均未命中 → 返回 entry 原值（支持 PATH 命令）。
func resolveEntrypoint(dir, entry string) (string, error) {
	if entry == "" {
		return "", fmt.Errorf("entrypoint 不能为空")
	}
	if filepath.IsAbs(entry) {
		return entry, nil
	}

	candidates := []string{filepath.Join(dir, entry)}
	if runtime.GOOS == "windows" && !strings.HasSuffix(strings.ToLower(entry), ".exe") {
		candidates = append(candidates, filepath.Join(dir, entry+".exe"))
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}
	return entry, nil
}
