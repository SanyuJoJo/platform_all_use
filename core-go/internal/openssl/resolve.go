package openssl

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "runtime"

    "github.com/yourorg/core-go/internal/platform"
)

var (
    coreRoot      string
    minVersion    string
    configuredBin string // 新增：来自 conf/core.yaml 的 tongsuo_openssl_bin
    resolvedBin   string
)

// Init 初始化 openssl 解析。
//
// 变更：新增 opensslBin 参数，来自 conf/core.yaml 的 tongsuo_openssl_bin。
func Init(root, opensslBin, minVer string) {
    coreRoot = root
    minVersion = minVer
    configuredBin = opensslBin
}

// Resolve 按平台解析 openssl 二进制路径。
//
// 解析优先级：
//   1. Init 传入的 configuredBin（conf/core.yaml 的 tongsuo_openssl_bin）
//   2. 环境变量 OPENSSL_BIN
//   3. coreRoot/libs/openssl/<os>-<arch>/openssl(.exe)
//   4. coreRoot/libs/bin/tongsuo/bin/openssl（兼容旧布局）
//   5. coreRoot/run/bin/openssl（兼容旧布局）
//   6. PATH 中的 openssl（兜底，但不保证是铜锁）
func Resolve() (string, error) {
    if resolvedBin != "" {
        return resolvedBin, nil
    }

    // 1. 配置文件的路径
    if configuredBin != "" {
        if info, err := os.Stat(configuredBin); err == nil && !info.IsDir() {
            resolvedBin = configuredBin
            return configuredBin, nil
        }
        // 配置了但不存在，继续尝试后续候选（不直接报错，便于迁移场景）
    }

    // 2. 环境变量
    if env := os.Getenv("OPENSSL_BIN"); env != "" {
        if _, err := os.Stat(env); err == nil {
            resolvedBin = env
            return env, nil
        }
    }

    // 3-5. 平台/兼容目录
    exeName := "openssl" + platform.ExeSuffix()
    candidates := []string{
        filepath.Join(coreRoot, "libs", "openssl",
            platform.OS()+"-"+platform.Arch(), exeName),
        filepath.Join(coreRoot, "libs", "bin", "tongsuo", "bin", exeName),
        filepath.Join(coreRoot, "run", "bin", exeName),
    }
    for _, c := range candidates {
        if info, err := os.Stat(c); err == nil && !info.IsDir() {
            resolvedBin = c
            return c, nil
        }
    }

    // 6. PATH 兜底
    if p, err := exec.LookPath("openssl"); err == nil {
        resolvedBin = p
        return p, nil
    }

    return "", fmt.Errorf("openssl not found; expected at %s",
        filepath.Join(coreRoot, "libs", "openssl",
            runtime.GOOS+"-"+runtime.GOARCH, exeName))
}
