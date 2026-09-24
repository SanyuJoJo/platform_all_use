package crypto
import (
	"fmt"
	"os"
	"path/filepath"
	"backend-go/internal/exception"
)
// FileStore 封装受控文件读写，防止越界访问。
type FileStore struct {
	coreRoot string
	guard    *PathGuard
}
func NewFileStore(coreRoot string) *FileStore {
	return &FileStore{coreRoot: coreRoot, guard: NewPathGuard(coreRoot)}
}
// ReadCoreFile 读取 data/ 或 tmp/ 下的文件。
func (s *FileStore) ReadCoreFile(rel string) ([]byte, error) {
	abs, err := s.guard.Resolve(rel, "path")
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, exception.New(
				exception.CodeNotFound, "文件不存在", 404, nil,
			)
		}
		return nil, exception.New(
			exception.CodeInternalError,
			fmt.Sprintf("读文件失败: %v", err), 500, nil,
		)
	}
	return data, nil
}
// WriteCoreFile 写入 data/ 或 tmp/ 下的文件，指定权限。
func (s *FileStore) WriteCoreFile(rel string, data []byte, mode os.FileMode) error {
	abs, err := s.guard.Resolve(rel, "path")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0750); err != nil {
		return exception.New(
			exception.CodeInternalError,
			fmt.Sprintf("创建目录失败: %v", err), 500, nil,
		)
	}
	if err := os.WriteFile(abs, data, mode); err != nil {
		return exception.New(
			exception.CodeInternalError,
			fmt.Sprintf("写文件失败: %v", err), 500, nil,
		)
	}
	return nil
}
// SafeRemove 静默删除文件（用于临时文件清理）。
func (s *FileStore) SafeRemove(rel string) error {
	abs, err := s.guard.Resolve(rel, "path")
	if err != nil {
		return err
	}
	return os.Remove(abs)
}
