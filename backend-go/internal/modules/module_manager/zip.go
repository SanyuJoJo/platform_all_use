package module_manager
import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"backend-go/internal/exception"
)
// SafeExtractZip 安全解压 ZIP 到 targetDir。
//
// 防护项：
//   - 路径穿越（memberPath 必须在 targetRoot 下）
//   - 符号链接
//   - 单文件大小 / 总大小 / 文件数量
//   - 压缩比炸弹
//
// v1.3（P1-REGRESSION-01）：
//   - 移除 v1.2 引入的 strings.ToLower 归一化；
//   - 直接使用 filepath.Rel 的原始语义，获取正确的跨平台行为：
//     · Windows：filepath.Rel 底层天然大小写不敏感；
//     · Linux/macOS：filepath.Rel 保持大小写敏感。
//   - 修复了 Linux/macOS 下 `/tmp/Module` 与 `/tmp/module` 被错误等价
//     导致 `../module/evil.txt` 绕过检查的安全漏洞。
func SafeExtractZip(
	zipPath, targetDir string,
	maxSize, maxTotal int64,
	maxFiles int,
) error {
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("创建解压目录失败：%v", err), 500, nil)
	}
	targetRoot, err := filepath.Abs(targetDir)
	if err != nil {
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("解析目标目录失败：%v", err), 500, nil)
	}
	targetRoot = filepath.Clean(targetRoot)
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return exception.New(exception.CodeModuleZipInvalid,
			fmt.Sprintf("打开 ZIP 失败：%v", err), 400, nil)
	}
	defer reader.Close()
	var totalSize int64
	fileCount := 0
	// 第一遍：全部校验
	for _, f := range reader.File {
		fileCount++
		if fileCount > maxFiles {
			return exception.New(exception.CodeModuleZipInvalid,
				fmt.Sprintf("ZIP 包文件数量超过上限 %d", maxFiles), 400, nil)
		}
		if int64(f.UncompressedSize64) > maxSize {
			return exception.New(exception.CodeModuleZipInvalid,
				fmt.Sprintf("ZIP 包内单文件超过大小上限 %d", maxSize), 400, nil)
		}
		if f.CompressedSize64 > 0 {
			ratio := float64(f.UncompressedSize64) / float64(f.CompressedSize64)
			if ratio > ZipMaxCompressRatio {
				return exception.New(exception.CodeModuleZipInvalid,
					fmt.Sprintf("ZIP 包内文件压缩比异常：%s", f.Name), 400, nil)
			}
		}
		totalSize += int64(f.UncompressedSize64)
		if totalSize > maxTotal {
			return exception.New(exception.CodeModuleZipInvalid,
				fmt.Sprintf("ZIP 包解压后总大小超过上限 %d", maxTotal), 400, nil)
		}
		// 符号链接检查
		if f.Mode()&os.ModeSymlink != 0 {
			return exception.New(exception.CodeModuleZipInvalid,
				fmt.Sprintf("ZIP 包包含符号链接：%s", f.Name), 400, nil)
		}
		// 路径穿越检查
		//
		// v1.3（P1-REGRESSION-01）：不做任何大小写归一化，
		// 直接使用 filepath.Rel 的原始语义。
		name := strings.ReplaceAll(f.Name, "\\", "/")
		memberPath := filepath.Join(targetRoot, name)
		memberPath, err = filepath.Abs(memberPath)
		if err != nil {
			return exception.New(exception.CodeModuleZipInvalid,
				fmt.Sprintf("ZIP 包路径解析失败：%s", f.Name), 400, nil)
		}
		memberPath = filepath.Clean(memberPath)
		rel, err := filepath.Rel(targetRoot, memberPath)
		if err != nil || strings.HasPrefix(rel, "..") {
			return exception.New(exception.CodeModuleZipInvalid,
				fmt.Sprintf("ZIP 包包含非法路径：%s", f.Name), 400, nil)
		}
	}
	// 第二遍：实际解压
	for _, f := range reader.File {
		if err := extractZipEntry(f, targetRoot); err != nil {
			return err
		}
	}
	return nil
}
// extractZipEntry 解压单个条目。
func extractZipEntry(f *zip.File, targetRoot string) error {
	name := strings.ReplaceAll(f.Name, "\\", "/")
	memberPath := filepath.Join(targetRoot, name)
	memberPath, _ = filepath.Abs(memberPath)
	memberPath = filepath.Clean(memberPath)
	if f.FileInfo().IsDir() {
		return os.MkdirAll(memberPath, 0o755)
	}
	if err := os.MkdirAll(filepath.Dir(memberPath), 0o755); err != nil {
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("创建目录失败：%v", err), 500, nil)
	}
	rc, err := f.Open()
	if err != nil {
		return exception.New(exception.CodeModuleZipInvalid,
			fmt.Sprintf("打开 ZIP 条目失败：%s", f.Name), 400, nil)
	}
	defer rc.Close()
	out, err := os.OpenFile(memberPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("创建文件失败：%v", err), 500, nil)
	}
	defer out.Close()
	if _, err := io.Copy(out, rc); err != nil {
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("写入文件失败：%v", err), 500, nil)
	}
	return nil
}
