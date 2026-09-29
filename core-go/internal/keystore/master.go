package keystore
import (
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "os"
    "path/filepath"
    "sync"
    "github.com/yourorg/core-go/internal/platform"
)
var (
    defaultStore *Store
    initOnce     sync.Once
    initErr      error
)
// Init 加载或生成主密钥。
//
//   - 若 masterKeyPath 存在 → 加载
//   - 否则 → 生成 32 字节随机密钥，以 hex 写入，权限 0600
func Init(masterKeyPath string) error {
    initOnce.Do(func() {
        if _, err := os.Stat(masterKeyPath); os.IsNotExist(err) {
            if err := generateMasterKey(masterKeyPath); err != nil {
                initErr = err
                return
            }
        }
        raw, err := os.ReadFile(masterKeyPath)
        if err != nil {
            initErr = fmt.Errorf("read master key: %w", err)
            return
        }
        defaultStore, initErr = NewStore(string(raw))
    })
    return initErr
}
// Default 返回全局 Store。
func Default() *Store {
    if defaultStore == nil {
        panic("keystore not initialized")
    }
    return defaultStore
}
func generateMasterKey(path string) error {
    dir := filepath.Dir(path)
    if err := platform.SecureMkdirAll(dir, 0700); err != nil {
        return fmt.Errorf("mkdir for master key: %w", err)
    }
    key := make([]byte, keyLen)
    if _, err := rand.Read(key); err != nil {
        return fmt.Errorf("generate random key: %w", err)
    }
    hexKey := hex.EncodeToString(key)
    if err := os.WriteFile(path, []byte(hexKey+"\n"), 0600); err != nil {
        return fmt.Errorf("write master key: %w", err)
    }
    _ = platform.SecureChmod(path, 0600)
    return nil
}
