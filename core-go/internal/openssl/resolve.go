package openssl

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
    "strings"

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
        p := configuredBin
        if !filepath.IsAbs(p) {
            p = filepath.Join(coreRoot, filepath.FromSlash(p))
        }
        if info, err := os.Stat(p); err == nil && !info.IsDir() {
            resolvedBin = p
            return p, nil
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
        filepath.Join(coreRoot, "libs", "tongsuo", "bin", exeName),
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


// childEnv 返回 openssl 子进程环境：把 core 包内铜锁 libs 目录注入
// LD_LIBRARY_PATH（/ DYLD_LIBRARY_PATH），使 openssl 在任意安装目录下
// 都能找到 libcrypto/libssl —— openssl 二进制的 RUNPATH 是构建机的绝对路径，
// 不能依赖它。
func childEnv() []string {
	env := os.Environ()

	// 1) 铜锁动态库路径（RUNPATH 是构建机绝对路径，必须注入）
	var dirs []string
	for _, d := range []string{
		filepath.Join(coreRoot, "libs", "tongsuo", "libs"),
		filepath.Join(coreRoot, "libs", "bin", "tongsuo", "libs"),
	} {
		if st, err := os.Stat(d); err == nil && st.IsDir() {
			dirs = append(dirs, d)
		}
	}
	if len(dirs) > 0 {
		key := "LD_LIBRARY_PATH"
		if runtime.GOOS == "darwin" {
			key = "DYLD_LIBRARY_PATH"
		}
		joined := strings.Join(dirs, string(os.PathListSeparator))
		if existing := os.Getenv(key); existing != "" {
			joined += string(os.PathListSeparator) + existing
		}
		env = append(env, key+"="+joined)
	}

	// 2) OPENSSL_CONF：openssl 二进制的 OPENSSLDIR 是编译期写死的构建机
	//    绝对路径（如 /code/.../tongsuo/ssl），换机后 openssl req 等命令
	//    找不到 openssl.cnf。这里显式指向 core 包内自带的 cnf。
	if os.Getenv("OPENSSL_CONF") == "" {
		for _, dir := range []string{
			filepath.Join(coreRoot, "libs", "tongsuo", "ssl"),
			filepath.Join(coreRoot, "libs", "bin", "tongsuo", "ssl"),
		} {
			cnf := filepath.Join(dir, "openssl.cnf")
			if st, err := os.Stat(cnf); err == nil && !st.IsDir() {
				env = append(env, "OPENSSL_CONF="+cnf)
				break
			}
		}
	}
	return env
}
