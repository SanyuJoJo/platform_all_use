package openssl

import (
    "fmt"
    "regexp"
    "strconv"
    "strings"
)

var (
    // 优先匹配 "Tongsuo x.y.z" 或 "Tongsuo: Tongsuo x.y.z"
    // (?i) 忽略大小写；[^0-9]* 跳过版本号之前的所有非数字字符
    tongsuoVerRe = regexp.MustCompile(`(?i)tongsuo[^0-9]*(\d+)\.(\d+)\.(\d+)`)

    // 通用匹配：任意 x.y.z
    genericVerRe = regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)
)

// parseVersion 从 version 输出中解析版本号。
//
// 解析策略：
//  1. 优先匹配 "Tongsuo x.y.z"——避免多行输出（Tongsuo + OpenSSL）时
//     误取到 OpenSSL 的版本号；
//  2. 回退到通用 x.y.z 匹配。
func parseVersion(raw string) (string, [3]int, error) {
    // 1. 优先 Tongsuo
    if m := tongsuoVerRe.FindStringSubmatch(raw); len(m) >= 4 {
        v := [3]int{}
        v[0], _ = strconv.Atoi(m[1])
        v[1], _ = strconv.Atoi(m[2])
        v[2], _ = strconv.Atoi(m[3])
        return fmt.Sprintf("%d.%d.%d", v[0], v[1], v[2]), v, nil
    }
    // 2. 通用匹配
    if m := genericVerRe.FindStringSubmatch(raw); len(m) >= 4 {
        v := [3]int{}
        v[0], _ = strconv.Atoi(m[1])
        v[1], _ = strconv.Atoi(m[2])
        v[2], _ = strconv.Atoi(m[3])
        return fmt.Sprintf("%d.%d.%d", v[0], v[1], v[2]), v, nil
    }
    return "", [3]int{}, fmt.Errorf("cannot parse version from: %s", raw)
}

// CheckVersion 校验铜锁版本 >= minVersion。
func CheckVersion() error {
    bin, err := Resolve()
    if err != nil {
        return err
    }
    ctx, cancel := contextWithTimeout()
    defer cancel()

    c := &Client{bin: bin}
    out, err := c.Run(ctx, "version")
    if err != nil {
        return fmt.Errorf("run openssl version: %w", err)
    }
    raw := strings.TrimSpace(string(out))

    gotStr, got, err := parseVersion(raw)
    if err != nil {
        return err
    }

    _, want, err := parseVersion("Tongsuo " + minVersion)
    if err != nil {
        return fmt.Errorf("invalid min version: %s", minVersion)
    }

    for i := 0; i < 3; i++ {
        if got[i] > want[i] {
            return nil
        }
        if got[i] < want[i] {
            return fmt.Errorf(
                "tongsuo version too low: %s < %s (parsed: %s)",
                raw, minVersion, gotStr)
        }
    }
    return nil
}

// Version 返回当前 openssl 版本字符串。
func Version() (string, error) {
    bin, err := Resolve()
    if err != nil {
        return "", err
    }
    ctx, cancel := contextWithTimeout()
    defer cancel()
    c := &Client{bin: bin}
    out, err := c.Run(ctx, "version")
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(out)), nil
}
