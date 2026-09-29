package config
import (
    "fmt"
    "os"
    "sync"
    "gopkg.in/yaml.v3"
)
// Whitelist 算法白名单（C-07）。
type Whitelist struct {
    SchemaVersion      string                       `yaml:"schema_version"`
    WhitelistVersion   string                       `yaml:"whitelist_version"`
    TongsuoMinVersion  string                       `yaml:"tongsuo_min_version"`
    Algorithms         map[string]AlgorithmSpec     `yaml:"algorithms"`
    Operations         map[string]OperationSpec     `yaml:"operations"`
}
type AlgorithmSpec struct {
    Type      []string `yaml:"type"`
    Curve     []string `yaml:"curve"`
    KeySize   []int    `yaml:"key_size"`
    Digest    []string `yaml:"digest"`
    Signature []string `yaml:"signature"`
    CertTypes []string `yaml:"cert_types"`
    Parameter []string `yaml:"parameter"`
    Priority  string   `yaml:"priority"`
}
type OperationSpec struct {
    AllowedAlgorithms        []string `yaml:"allowed_algorithms"`
    AllowedCertTypes         []string `yaml:"allowed_cert_types"`
    AllowedCAKinds           []string `yaml:"allowed_ca_source"`
    AllowedDigestAlgorithms  []string `yaml:"allowed_digest_algorithms"`
}
var (
    whitelistCache *Whitelist
    whitelistOnce  sync.Once
    whitelistErr   error
)
// LoadWhitelist 加载算法白名单。
func LoadWhitelist(path string) (*Whitelist, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("read whitelist: %w", err)
    }
    var w Whitelist
    if err := yaml.Unmarshal(data, &w); err != nil {
        return nil, fmt.Errorf("parse whitelist: %w", err)
    }
    return &w, nil
}
// GetWhitelist 返回单例（首次加载）。
func GetWhitelist(path string) (*Whitelist, error) {
    whitelistOnce.Do(func() {
        whitelistCache, whitelistErr = LoadWhitelist(path)
    })
    return whitelistCache, whitelistErr
}
// IsAlgorithmAllowed 判断 operation 是否允许某算法。
func (w *Whitelist) IsAlgorithmAllowed(op, alg string) bool {
    if w == nil {
        return false
    }
    spec, ok := w.Operations[op]
    if !ok {
        return false
    }
    for _, a := range spec.AllowedAlgorithms {
        if a == "*" || a == alg {
            return true
        }
    }
    return false
}
// IsCertTypeAllowed 判断 operation 是否允许某证书类型。
func (w *Whitelist) IsCertTypeAllowed(op, certType string) bool {
    if w == nil {
        return false
    }
    spec, ok := w.Operations[op]
    if !ok {
        return false
    }
    if len(spec.AllowedCertTypes) == 0 {
        return true // 未声明则不限
    }
    for _, t := range spec.AllowedCertTypes {
        if t == certType {
            return true
        }
    }
    return false
}
// IsCAKindAllowed 判断 operation 是否允许某种 CA 来源。
func (w *Whitelist) IsCAKindAllowed(op, caKind string) bool {
    if w == nil {
        return false
    }
    spec, ok := w.Operations[op]
    if !ok {
        return false
    }
    if len(spec.AllowedCAKinds) == 0 {
        return true
    }
    for _, k := range spec.AllowedCAKinds {
        if k == caKind {
            return true
        }
    }
    return false
}
