package handler

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// writeRehashLink 在证书同目录下创建 OpenSSL c_rehash 兼容的软链：
//
//	<dir>/<hash>.0 → <cert_basename>
//
// 用途：让 `openssl verify -CApath <dir>` 能自动找到该 CA。
//
// 行为：
//  1. 清理本目录下指向不存在文件的孤儿 .0 软链；
//  2. 计算本证书的新版 subject_hash（-subject_hash）；
//  3. 建 <hash>.0 → <basename> 软链；已存在则覆盖。
//
// ★ 只生成新版 hash：
//   - 铜锁基于 OpenSSL 1.1.1+，默认使用新版 subject_hash；
//   - 现代客户端也使用新版算法；
//   - 不再生成旧版 hash（-subject_hash_old），避免每个 CA 两个软链。
func writeRehashLink(certAbs string) error {
	if certAbs == "" {
		return fmt.Errorf("cert path is empty")
	}
	dir := filepath.Dir(certAbs)
	base := filepath.Base(certAbs)

	// 1. 清理孤儿 rehash 软链
	cleanupOrphanRehashLinks(dir)

	// 2. 找 openssl
	opensslBin, err := findOpenSSLForRehash()
	if err != nil {
		return err
	}

	// 3. 计算新版 subject_hash
	hash, err := computeSubjectHash(opensslBin, certAbs, "")
	if err != nil {
		return fmt.Errorf("subject_hash: %w", err)
	}
	if hash == "" {
		return fmt.Errorf("empty subject_hash for %s", certAbs)
	}

	// 4. 建 <hash>.0 → <basename> 软链
	link := filepath.Join(dir, hash+".0")
	_ = os.Remove(link)
	if err := os.Symlink(base, link); err != nil {
		return fmt.Errorf("symlink %s → %s: %w", link, base, err)
	}

	return nil
}

// cleanupOrphanRehashLinks 清理 dir 下所有指向不存在文件的 .0 软链。
//
// 场景：
//   - CA 被删除后，rehash 软链未被清理（旧代码遗留）；
//   - 手工删除了 .cert.pem 但没删 .0；
//   - 并发删除 + 重建导致的残留。
//
// 只清理软链，不动真实文件。
func cleanupOrphanRehashLinks(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".0") || len(name) != 10 {
			continue
		}
		full := filepath.Join(dir, name)
		info, err := os.Lstat(full)
		if err != nil || info.Mode()&os.ModeSymlink == 0 {
			continue
		}
		target, err := os.Readlink(full)
		if err != nil {
			continue
		}
		// 相对路径 → 相对 dir 解析
		absTarget := target
		if !filepath.IsAbs(absTarget) {
			absTarget = filepath.Join(dir, target)
		}
		if _, err := os.Stat(absTarget); err != nil {
			// 目标不存在 → 孤儿软链
			_ = os.Remove(full)
		}
	}
}

// findOpenSSLForRehash 找一个可用的 openssl 二进制。
//
// 优先铜锁（与签名算法一致），回退系统 openssl。
func findOpenSSLForRehash() (string, error) {
	// 从环境变量拿铜锁路径（main.go 里已初始化）
	if bin := os.Getenv("PLATS_OPENSSL_BIN"); bin != "" {
		if _, err := os.Stat(bin); err == nil {
			return bin, nil
		}
	}
	// 回退系统 openssl
	if bin, err := exec.LookPath("openssl"); err == nil {
		return bin, nil
	}
	return "", fmt.Errorf("openssl not found")
}

// computeSubjectHash 计算证书 subject hash。
//
// flag 为 "" 时用 -subject_hash；为 "-subject_hash_old" 时用旧算法。
func computeSubjectHash(opensslBin, certAbs, flag string) (string, error) {
	args := []string{"x509", "-in", certAbs, "-noout"}
	if flag != "" {
		args = append(args, flag)
	} else {
		args = append(args, "-subject_hash")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, opensslBin, args...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
