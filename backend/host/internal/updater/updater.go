package updater
import (
    "archive/tar"
    "compress/gzip"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "io"
    "log"
    "os"
    "path/filepath"
    "sort"
    "time"
)
// Updater 负责主应用二进制的升级与回退
type Updater struct {
    binaryPath string
    backupDir  string
    maxBackups int
}
func NewUpdater(binaryPath, backupDir string) *Updater {
    _ = os.MkdirAll(backupDir, 0o755)
    return &Updater{binaryPath: binaryPath, backupDir: backupDir, maxBackups: 5}
}
// Snapshot 升级前创建快照
func (u *Updater) Snapshot(version string) (string, error) {
    ts := time.Now().Format("20060102_150405")
    dir := filepath.Join(u.backupDir, version+"_"+ts)
    if err := os.MkdirAll(dir, 0o755); err != nil {
        return "", err
    }
    dst := filepath.Join(dir, filepath.Base(u.binaryPath))
    if err := copyFile(u.binaryPath, dst); err != nil {
        return "", err
    }
    log.Printf("[updater] 快照创建: %s", dir)
    u.cleanOld()
    return dir, nil
}
// ApplyUpdate 应用 tar.gz 升级包
func (u *Updater) ApplyUpdate(tarPath, expectedChecksum string) error {
    if err := verifyFileChecksum(tarPath, expectedChecksum); err != nil {
        return fmt.Errorf("升级包校验失败: %w", err)
    }
    tmpDir, err := os.MkdirTemp("", "update-*")
    if err != nil {
        return err
    }
    defer os.RemoveAll(tmpDir)
    if err := untarGz(tarPath, tmpDir); err != nil {
        return fmt.Errorf("解压失败: %w", err)
    }
    newBinary := filepath.Join(tmpDir, filepath.Base(u.binaryPath))
    if _, err := os.Stat(newBinary); err != nil {
        return fmt.Errorf("升级包缺少主程序: %w", err)
    }
    old := u.binaryPath + ".old"
    if err := os.Rename(u.binaryPath, old); err != nil {
        return err
    }
    if err := copyFile(newBinary, u.binaryPath); err != nil {
        _ = os.Rename(old, u.binaryPath)
        return fmt.Errorf("替换失败: %w", err)
    }
    _ = os.Remove(old)
    log.Println("[updater] 升级完成")
    return nil
}
// Rollback 回退到指定快照目录
func (u *Updater) Rollback(snapshotDir string) error {
    entries, err := os.ReadDir(snapshotDir)
    if err != nil {
        return err
    }
    for _, e := range entries {
        src := filepath.Join(snapshotDir, e.Name())
        dst := filepath.Join(filepath.Dir(u.binaryPath), e.Name())
        if err := copyFile(src, dst); err != nil {
            return err
        }
    }
    log.Printf("[updater] 已回退到 %s", snapshotDir)
    return nil
}
// ListSnapshots 列出全部快照（按时间倒序）
func (u *Updater) ListSnapshots() ([]string, error) {
    entries, err := os.ReadDir(u.backupDir)
    if err != nil {
        return nil, err
    }
    var dirs []string
    for _, e := range entries {
        if e.IsDir() {
            dirs = append(dirs, filepath.Join(u.backupDir, e.Name()))
        }
    }
    sort.Sort(sort.Reverse(sort.StringSlice(dirs)))
    return dirs, nil
}
func (u *Updater) cleanOld() {
    dirs, _ := u.ListSnapshots()
    for len(dirs) > u.maxBackups {
        _ = os.RemoveAll(dirs[len(dirs)-1])
        dirs = dirs[:len(dirs)-1]
    }
}
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
func verifyFileChecksum(path, expected string) error {
    if expected == "" {
        return nil
    }
    data, err := os.ReadFile(path)
    if err != nil {
        return err
    }
    h := sha256.Sum256(data)
    if hex.EncodeToString(h[:]) != expected {
        return fmt.Errorf("校验和不匹配")
    }
    return nil
}
func untarGz(src, dst string) error {
    f, err := os.Open(src)
    if err != nil {
        return err
    }
    defer f.Close()
    gz, err := gzip.NewReader(f)
    if err != nil {
        return err
    }
    defer gz.Close()
    tr := tar.NewReader(gz)
    for {
        header, err := tr.Next()
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }
        target := filepath.Join(dst, header.Name)
        switch header.Typeflag {
        case tar.TypeReg:
            if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
                return err
            }
            out, err := os.Create(target)
            if err != nil {
                return err
            }
            if _, err := io.Copy(out, tr); err != nil {
                out.Close()
                return err
            }
            out.Close()
        case tar.TypeDir:
            _ = os.MkdirAll(target, 0o755)
        }
    }
    return nil
}
