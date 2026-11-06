package crypto

import (
	"bytes"
	"context"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"backend-go/internal/exception"
)

// =============================================================================
// CertParser / CertDetail
// =============================================================================

// CertParser 证书解析。优先铜锁 openssl，Go 标准库兜底。
type CertParser struct {
	coreRoot string
}

func NewCertParser(coreRoot string) *CertParser {
	return &CertParser{coreRoot: coreRoot}
}

// CertDetail 证书解析结果。
type CertDetail struct {
	Version            string `json:"version"`
	Serial             string `json:"serial"`
	Issuer             string `json:"issuer"`
	Subject            string `json:"subject"`
	Fingerprint        string `json:"fingerprint"`
	NotBefore          string `json:"not_before"`
	NotAfter           string `json:"not_after"`
	PublicKeyAlgorithm string `json:"public_key_algorithm"`
	SignatureAlgorithm string `json:"signature_algorithm"`
	SignatureValue     string `json:"signature_value"`
	PublicKeyValue     string `json:"public_key_value"`
	KeyUsage           string `json:"key_usage"`
	ExtendedKeyUsage   string `json:"extended_key_usage"`
}

// CADetail 是 CertDetail 的类型别名，保留以兼容已有 CA 代码。
type CADetail = CertDetail

// ★ 展示时区统一为北京时间 UTC+8（与第三方平台对齐）
var cstZone = time.FixedZone("CST", 8*3600)

// =============================================================================
// OID → 展示名映射
// =============================================================================
//
// 覆盖三类 OID：
//  1. 标准 OID（SM2 / RSA / EC / ML-DSA / SLH-DSA）；
//  2. OQS 私用 OID（1.3.9999.3.x），标准 oqs-provider 表；
//  3. 现场定制 oqs-provider 的特殊映射 → 通过 CRYPTO_OID_<oid> 环境变量覆盖。
//
// ★ Falcon 特殊说明：
//
//	标准 oqs-provider：falcon512=1.3.9999.3.7 / falcon1024=1.3.9999.3.8
//	现场定制版本可能与标准不同（例如把 falcon1024 编到了 1.3.9999.3.14）→
//	Parse 阶段会用 `openssl pkey -text` 动态探测优先级高于本表；
//	如果动态探测也失败，通过 CRYPTO_OID_<oid> 环境变量兜底。
var oidToDisplay = map[string]string{
	// ---- 标准 OID ----
	"1.2.156.10197.1.301":     "ECC(256)",           // SM2
	"1.2.840.113549.1.1.1":    "RSA",                // rsaEncryption
	"1.2.840.10045.2.1":       "ECC(256)",           // id-ecPublicKey（曲线靠 ASN1 OID 识别，默认 256）
	"2.16.840.1.101.3.4.3.17": "ML-DSA-44",          // FIPS 204
	"2.16.840.1.101.3.4.3.18": "ML-DSA-65",
	"2.16.840.1.101.3.4.3.19": "ML-DSA-87",
	"2.16.840.1.101.3.4.3.20": "SLH-DSA-SHA2-128s",  // FIPS 205
	"2.16.840.1.101.3.4.3.21": "SLH-DSA-SHA2-128f",
	"2.16.840.1.101.3.4.3.22": "SLH-DSA-SHA2-192s",
	"2.16.840.1.101.3.4.3.23": "SLH-DSA-SHA2-192f",
	"2.16.840.1.101.3.4.3.24": "SLH-DSA-SHA2-256s",
	"2.16.840.1.101.3.4.3.25": "SLH-DSA-SHA2-256f",
	"2.16.840.1.101.3.4.3.26": "SLH-DSA-SHAKE-128s",
	"2.16.840.1.101.3.4.3.27": "SLH-DSA-SHAKE-128f",
	"2.16.840.1.101.3.4.3.28": "SLH-DSA-SHAKE-192s",
	"2.16.840.1.101.3.4.3.29": "SLH-DSA-SHAKE-192f",
	"2.16.840.1.101.3.4.3.30": "SLH-DSA-SHAKE-256s",
	"2.16.840.1.101.3.4.3.31": "SLH-DSA-SHAKE-256f",

	// ---- OQS 私用 OID（标准 oqs-provider 表）----
	"1.3.9999.3.1":  "ML-DSA-44",           // dilithium2
	"1.3.9999.3.2":  "ML-DSA-65",           // dilithium3
	"1.3.9999.3.3":  "ML-DSA-87",           // dilithium5
	"1.3.9999.3.4":  "ML-DSA-44-AES",
	"1.3.9999.3.5":  "ML-DSA-65-AES",
	"1.3.9999.3.6":  "ML-DSA-87-AES",
	"1.3.9999.3.7":  "Falcon(512)",
	"1.3.9999.3.8":  "Falcon(1024)",
	"1.3.9999.3.9":  "Falcon-Padded-512",
	"1.3.9999.3.10": "Falcon-Padded-1024",
	"1.3.9999.3.11": "SLH-DSA-SHA2-128f",   // sphincssha2128fsimple
	"1.3.9999.3.12": "SLH-DSA-SHA2-128s",
	"1.3.9999.3.13": "SLH-DSA-SHA2-192f",
	"1.3.9999.3.14": "SLH-DSA-SHA2-192s",
	"1.3.9999.3.15": "SLH-DSA-SHA2-256f",
	"1.3.9999.3.16": "SLH-DSA-SHA2-256s",
	"1.3.9999.3.17": "SLH-DSA-SHAKE-128f",
	"1.3.9999.3.18": "SLH-DSA-SHAKE-128s",
	"1.3.9999.3.19": "SLH-DSA-SHAKE-192f",
	"1.3.9999.3.20": "SLH-DSA-SHAKE-192s",
	"1.3.9999.3.21": "SLH-DSA-SHAKE-256f",
	"1.3.9999.3.22": "SLH-DSA-SHAKE-256s",
}

// oidOverride 环境变量覆盖某个 OID 的展示名（终极兜底）。
//
//	CRYPTO_OID_1.3.9999.3.14=Falcon(1024)
func oidOverride(oid string) string {
	if v := strings.TrimSpace(os.Getenv("CRYPTO_OID_" + oid)); v != "" {
		return v
	}
	return ""
}

// lookupOIDDisplay 按 OID 查展示名；找不到返回空。
func lookupOIDDisplay(oid string) string {
	oid = strings.TrimSpace(oid)
	if oid == "" {
		return ""
	}
	if v := oidOverride(oid); v != "" {
		return v
	}
	if v, ok := oidToDisplay[oid]; ok {
		return v
	}
	// OQS 段但表中未收录 → 给可读占位，避免前端展示裸 OID
	if strings.HasPrefix(oid, "1.3.9999.3.") {
		return "OQS-" + oid
	}
	return ""
}

// =============================================================================
// Parse
// =============================================================================

// Parse 解析证书。
//
// ★ 解析策略（同时用 openssl + Go 标准库，取长补短）：
//
//  1. openssl 优先：完整解析所有字段；
//  2. Go 标准库补充：SPKI / 签名 DER / 精确时间；
//  3. ★ OID 动态探测：如果公钥算法是裸 OID（1.3.9999.3.x），
//     通过 `openssl pkey -text` 从公钥内容获取 provider 返回的真实算法名；
//     探测失败时回退到 oidToDisplay 硬编码表 + 环境变量覆盖；
//  4. 展示名归一化：把 openssl 的 OID 原始名映射为第三方平台展示名；
//  5. 与第三方平台对齐：摘要值 SHA1、时间 UTC+8、DN 顺序、SPKI hex。
func (p *CertParser) Parse(certAbs string) (*CertDetail, error) {
	bin := p.OpensslBin()

	// 1. openssl 完整解析
	var detail *CertDetail
	if bin != "" {
		if d, err := p.parseWithOpenSSL(bin, certAbs); err == nil && d.Subject != "" {
			detail = d
		}
	}

	// 2. Go 标准库补充（SPKI / 签名 DER / 精确时间）
	if goD, err := p.parseFromPEM(certAbs); err == nil && goD != nil {
		if detail == nil {
			detail = goD
		} else {
			if goD.PublicKeyValue != "" {
				detail.PublicKeyValue = goD.PublicKeyValue
			}
			if goD.SignatureValue != "" {
				detail.SignatureValue = goD.SignatureValue
			}
		}
	}

	if detail == nil || detail.Subject == "" {
		if _, err := p.parseFromPEM(certAbs); err != nil {
			return nil, err
		}
		return nil, exception.New(
			exception.CodeParamInvalid, "证书解析失败", 400, nil,
		)
	}

	// 3. ★ OID 动态探测（关键：解决 oqs-provider 版本差异导致的映射错位）
	//
	//    openssl x509 -text 对 oqs-provider 算法输出裸 OID（1.3.9999.3.x），
	//    因为 openssl 的 OBJ 名字库只覆盖标准算法。
	//    但 openssl pkey -text 走 provider，provider 会返回真实名字。
	//    因此：公钥算法是裸 OID 时，用 pkey -text 探一次。
	if bin != "" && isPureOID(detail.PublicKeyAlgorithm) {
		if name := detectKeyAlgNameFromCert(bin, certAbs); name != "" {
			detail.PublicKeyAlgorithm = name
		}
	}

	// 4. ★ 统一展示名（此时 detail.PublicKeyAlgorithm 要么已经是真名，要么还是 OID）
	detail.PublicKeyAlgorithm = normalizePublicKeyAlg(detail.PublicKeyAlgorithm)
	detail.SignatureAlgorithm = normalizeSignatureAlg(detail.SignatureAlgorithm)

	// 5. 补全 openssl 解析的兜底字段
	if detail.Serial == "" && bin != "" {
		if serialOut, err := RunOpenSSL(bin, "x509", "-in", certAbs, "-noout", "-serial"); err == nil {
			s := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(serialOut), "serial="))
			detail.Serial = strings.ToLower(cleanHex(s))
		}
	}
	if detail.KeyUsage == "" && bin != "" {
		if kuOut, err := RunOpenSSL(bin, "x509", "-in", certAbs, "-noout", "-ext", "keyUsage"); err == nil {
			if lines := extractExtLines(kuOut); len(lines) > 0 {
				detail.KeyUsage = strings.Join(lines, ", ")
			}
		}
	}
	if detail.ExtendedKeyUsage == "" && bin != "" {
		if ekuOut, err := RunOpenSSL(bin, "x509", "-in", certAbs, "-noout", "-ext", "extendedKeyUsage"); err == nil {
			if lines := extractExtLines(ekuOut); len(lines) > 0 {
				detail.ExtendedKeyUsage = strings.Join(lines, ", ")
			}
		}
	}

	// 6. DN 顺序统一为 C → ST → L → O → OU → CN
	detail.Issuer = normalizeDNOrder(detail.Issuer)
	detail.Subject = normalizeDNOrder(detail.Subject)

	// 7. 公钥值统一为完整 SPKI DER hex（字节间空格分隔）
	if bin != "" {
		if spkiHex, err := p.extractSPKIHex(bin, certAbs); err == nil && spkiHex != "" {
			detail.PublicKeyValue = spkiHex
		}
	}

	return detail, nil
}

// OpensslBin 返回铜锁 openssl 路径。
func (p *CertParser) OpensslBin() string {
	candidates := []string{
		filepath.Join(p.coreRoot, "libs", "tongsuo", "bin", "openssl"),
		filepath.Join(p.coreRoot, "libs", "bin", "tongsuo", "bin", "openssl"),
		filepath.Join(p.coreRoot, "run", "bin", "openssl"),
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && !st.IsDir() && st.Mode()&0111 != 0 {
			return c
		}
	}
	if bin, err := exec.LookPath("openssl"); err == nil {
		return bin
	}
	return ""
}

// =============================================================================
// ★ OID 动态探测：从公钥内容获取 provider 返回的真实算法名
// =============================================================================

// detectKeyAlgNameFromCert 用 `openssl pkey -text` 从公钥内容获取 provider 返回的真名。
//
// 为什么需要：
//
//	openssl x509 -text 输出 Public Key Algorithm 依赖 libcrypto 的 OBJ 名字库，
//	该库只覆盖标准算法（RSA/SM2/ECC/ML-DSA/SLH-DSA 等由铜锁内置的算法）。
//	oqs-provider 提供的 Falcon 等算法不在 OBJ 库里 → 输出裸 OID（1.3.9999.3.x）。
//
//	openssl pkey -text 走 provider（OSSL），provider 内部知道自己注册了什么名字，
//	因此输出的第一行就是真实算法名（如 "Falcon-1024 Public-Key:"）。
//
// 兜底策略：
//
//	1. 先尝试默认加载（openssl.cnf 配置的 provider）；
//	2. 失败时显式加载 oqsprovider + default；
//	3. 再失败时尝试 oqs + default（兼容自编译版本）。
func detectKeyAlgNameFromCert(bin, certAbs string) string {
	if bin == "" || certAbs == "" {
		return ""
	}

	// 提取公钥 PEM
	pubPEM, err := RunOpenSSL(bin, "x509", "-in", certAbs, "-pubkey", "-noout")
	if err != nil || strings.TrimSpace(pubPEM) == "" {
		return ""
	}

	tmpDir, err := os.MkdirTemp("", "cert-alg-")
	if err != nil {
		return ""
	}
	defer os.RemoveAll(tmpDir)

	pubPath := filepath.Join(tmpDir, "pub.pem")
	if err := os.WriteFile(pubPath, []byte(pubPEM), 0600); err != nil {
		return ""
	}

	providers := [][]string{
		{},                        // 默认加载
		{"oqsprovider", "default"}, // 标准 oqs-provider 名
		{"oqs", "default"},         // 自编译版本名
	}
	for _, pv := range providers {
		args := []string{"pkey", "-pubin", "-in", pubPath, "-noout", "-text"}
		for _, p := range pv {
			args = append(args, "-provider", p)
		}
		out, err := RunOpenSSL(bin, args...)
		if err != nil {
			continue
		}
		if name := extractKeyAlgName(out); name != "" {
			return name
		}
	}
	return ""
}

// extractKeyAlgName 从 pkey -text 输出中提取算法名。
//
// 输出格式（首行）：
//
//	Falcon-1024 Public-Key:
//	    pub:
//	    ...
//	ML-DSA-65 Public-Key:
//	    ...
//	Falcon-1024 Private-Key:
//	    priv:
//	    ...
//
// 返回 "Falcon-1024" / "ML-DSA-65" 等；找不到返回空字符串。
func extractKeyAlgName(text string) string {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// 逐个匹配常见后缀（兼容不同 openssl 版本）
		for _, suffix := range []string{
			" Public-Key:",
			" Private-Key:",
			" Public Key:",
			" Private Key:",
			" public key:",
		} {
			if strings.HasSuffix(line, suffix) {
				name := strings.TrimSpace(strings.TrimSuffix(line, suffix))
				if name != "" {
					return name
				}
			}
		}
		// 遇到第一个非 "xxx:" 的行说明不是算法行，停止
		if !strings.HasSuffix(line, ":") {
			break
		}
	}
	return ""
}

// =============================================================================
// certAlgorithmID —— 从证书 PEM 推断规范算法 ID
// =============================================================================
//
// 用于 cert_import.go / ca_import.go 落库前填充 Algorithm 字段：
// 必须存**规范 ID**（"SM2" / "RSA2048" / "falcon1024" / "MLDSA65"），
// 而不是解析展示名（"ECC(256)" / "Falcon(1024)"）。
//
// 推断顺序：
//  1. openssl x509 -text 的 "Public Key Algorithm:" 行（标准算法名或裸 OID）；
//  2. 裸 OID 时查 oidToDisplay 反查规范 ID；
//  3. OID 表未命中时用 pkey -text 动态探测；
//  4. 动态探测也失败时返回空（调用方兜底）。
func certAlgorithmID(opensslBin, certAbs string) string {
	if opensslBin == "" || certAbs == "" {
		return ""
	}

	textOut, err := RunOpenSSL(opensslBin, "x509", "-in", certAbs, "-noout", "-text", "-nameopt", "RFC2253")
	if err != nil {
		return ""
	}

	// 1. 尝试 "Public Key Algorithm:" 行
	algName := extractPubAlgFromX509Text(textOut)

	// 2. 裸 OID → 查表
	if isPureOID(algName) {
		if disp := lookupOIDDisplay(algName); disp != "" {
			if id := displayNameToID(disp, textOut); id != "" {
				return id
			}
		}
		// 3. 表未命中 → 动态探测
		if name := detectKeyAlgNameFromCert(opensslBin, certAbs); name != "" {
			algName = name
		}
	}

	// 4. 名字 → 规范 ID
	return displayNameToID(algName, textOut)
}

// extractPubAlgFromX509Text 从 x509 -text 中提取 "Public Key Algorithm:" 行。
func extractPubAlgFromX509Text(text string) string {
	for _, line := range strings.Split(text, "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "Public Key Algorithm:") {
			return strings.TrimSpace(strings.TrimPrefix(t, "Public Key Algorithm:"))
		}
	}
	return ""
}

// extractPubAlgFromCSRText 从 `openssl req -text` 输出中提取公钥算法名。
//
// 与 extractPubAlgFromX509Text 相同逻辑（`openssl req -text` 也输出
// `Public Key Algorithm:` 行），单独函数仅为语义清晰，供 csr_service.go 调用。
func extractPubAlgFromCSRText(text string) string {
	return extractPubAlgFromX509Text(text)
}

// displayNameToID 把算法名或展示名转为规范 ID。
//
//	"SM2" / "ECC(256)"                              → "SM2"
//	"RSA" + text 含 "(2048 bit)"                     → "RSA2048"
//	"Falcon(512)" / "Falcon-512" / "falcon512"       → "falcon512"
//	"ML-DSA-65" / "mldsa65" / "dilithium3"           → "MLDSA65"
//	"SLH-DSA-SHA2-128s"                              → "slh-dsa-sha2-128s"
func displayNameToID(name, x509Text string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	if s == "" {
		return ""
	}

	switch {
	// ---- SM2 / ECC ----
	case strings.Contains(s, "sm2"),
		strings.Contains(s, "ecc"),
		strings.Contains(s, "id-ecpublickey"),
		strings.Contains(s, "prime256v1"):
		return "SM2"

	// ---- RSA ----
	case strings.Contains(s, "rsa"):
		switch {
		case strings.Contains(x509Text, "(4096 bit)"):
			return "RSA4096"
		case strings.Contains(x509Text, "(2048 bit)"):
			return "RSA2048"
		case strings.Contains(x509Text, "(1024 bit)"):
			return "RSA1024"
		}
		return "RSA2048"

	// ---- ML-DSA ----
	case strings.Contains(s, "ml-dsa-44"), strings.Contains(s, "mldsa44"), strings.Contains(s, "dilithium2"):
		return "MLDSA44"
	case strings.Contains(s, "ml-dsa-65"), strings.Contains(s, "mldsa65"), strings.Contains(s, "dilithium3"):
		return "MLDSA65"
	case strings.Contains(s, "ml-dsa-87"), strings.Contains(s, "mldsa87"), strings.Contains(s, "dilithium5"):
		return "MLDSA87"

	// ---- Falcon ----
	case strings.Contains(s, "falcon(512)"),
		strings.Contains(s, "falcon-512"),
		strings.Contains(s, "falcon512"):
		return "falcon512"
	case strings.Contains(s, "falcon(1024)"),
		strings.Contains(s, "falcon-1024"),
		strings.Contains(s, "falcon1024"):
		return "falcon1024"
	}

	// ---- SLH-DSA（全 12 变体）----
	for _, h := range []string{"sha2", "shake"} {
		for _, lvl := range []string{"128", "192", "256"} {
			for _, spd := range []string{"s", "f"} {
				id := "slh-dsa-" + h + "-" + lvl + spd
				if strings.Contains(s, id) ||
					strings.Contains(s, strings.ReplaceAll(id, "-", "")) {
					return id
				}
			}
		}
	}

	return ""
}

// =============================================================================
// 展示名归一化
// =============================================================================

// normalizePublicKeyAlg 归一化公钥算法展示名。
//
// ★ 匹配优先级：
//  1. OID 表（1.3.9999.3.x / 标准 OID）→ 直接返回展示名；
//  2. 名字模糊匹配（SM2 / RSA / ML-DSA / Falcon / SLH-DSA / 传统算法）；
//  3. 裸 OID 兜底 → "OID-<oid>"。
//
// ★ 注意：Falcon 返回**括号形态**（Falcon(512) / Falcon(1024)），
// 与 algorithms.go 中 AlgorithmSpec.DisplayName 完全一致。
func normalizePublicKeyAlg(alg string) string {
	s := strings.TrimSpace(alg)
	if s == "" {
		return ""
	}

	// 1. OID 优先
	if disp := lookupOIDDisplay(s); disp != "" {
		return disp
	}

	lower := strings.ToLower(s)

	switch {
	// ---- SM2 / ECC ----
	case strings.Contains(lower, "id-ecpublickey"),
		strings.Contains(lower, "ecdsa"),
		strings.Contains(lower, "sm2"),
		strings.Contains(lower, "ecc"),
		strings.Contains(lower, "prime256v1"),
		strings.Contains(lower, "secp256"):
		return "ECC(256)"
	case strings.Contains(lower, "secp384"):
		return "ECC(384)"
	case strings.Contains(lower, "secp521"):
		return "ECC(521)"

	// ---- RSA ----
	case strings.Contains(lower, "rsaencryption"),
		strings.Contains(lower, "rsa"):
		return "RSA"

	// ---- ML-DSA ----
	case strings.Contains(lower, "ml-dsa-44"),
		strings.Contains(lower, "mldsa44"),
		strings.Contains(lower, "dilithium2"):
		return "ML-DSA-44"
	case strings.Contains(lower, "ml-dsa-65"),
		strings.Contains(lower, "mldsa65"),
		strings.Contains(lower, "dilithium3"):
		return "ML-DSA-65"
	case strings.Contains(lower, "ml-dsa-87"),
		strings.Contains(lower, "mldsa87"),
		strings.Contains(lower, "dilithium5"):
		return "ML-DSA-87"

	// ---- Falcon ----
	case strings.Contains(lower, "falcon(512)"),
		strings.Contains(lower, "falcon-512"),
		strings.Contains(lower, "falcon512"):
		return "Falcon(512)"
	case strings.Contains(lower, "falcon(1024)"),
		strings.Contains(lower, "falcon-1024"),
		strings.Contains(lower, "falcon1024"):
		return "Falcon(1024)"

	// ---- SLH-DSA ----
	case strings.Contains(lower, "slh-dsa-sha2-128s"):
		return "SLH-DSA-SHA2-128s"
	case strings.Contains(lower, "slh-dsa-sha2-128f"):
		return "SLH-DSA-SHA2-128f"
	case strings.Contains(lower, "slh-dsa-sha2-192s"):
		return "SLH-DSA-SHA2-192s"
	case strings.Contains(lower, "slh-dsa-sha2-192f"):
		return "SLH-DSA-SHA2-192f"
	case strings.Contains(lower, "slh-dsa-sha2-256s"):
		return "SLH-DSA-SHA2-256s"
	case strings.Contains(lower, "slh-dsa-sha2-256f"):
		return "SLH-DSA-SHA2-256f"
	case strings.Contains(lower, "slh-dsa-shake-128s"):
		return "SLH-DSA-SHAKE-128s"
	case strings.Contains(lower, "slh-dsa-shake-128f"):
		return "SLH-DSA-SHAKE-128f"
	case strings.Contains(lower, "slh-dsa-shake-192s"):
		return "SLH-DSA-SHAKE-192s"
	case strings.Contains(lower, "slh-dsa-shake-192f"):
		return "SLH-DSA-SHAKE-192f"
	case strings.Contains(lower, "slh-dsa-shake-256s"):
		return "SLH-DSA-SHAKE-256s"
	case strings.Contains(lower, "slh-dsa-shake-256f"):
		return "SLH-DSA-SHAKE-256f"

	// ---- 其它传统算法 ----
	case strings.Contains(lower, "ed25519"):
		return "Ed25519"
	case strings.Contains(lower, "ed448"):
		return "Ed448"
	case strings.Contains(lower, "dsa"):
		return "DSA"
	}

	// 3. 裸 OID 兜底
	if isPureOID(s) {
		return "OID-" + s
	}
	return s
}

// normalizeSignatureAlg 归一化签名算法展示名。
func normalizeSignatureAlg(alg string) string {
	s := strings.TrimSpace(alg)
	if s == "" {
		return ""
	}

	// 1. OID 优先
	if disp := lookupOIDDisplay(s); disp != "" {
		return disp
	}

	lower := strings.ToLower(s)

	switch {
	// ---- SM2 ----
	case strings.Contains(lower, "sm2-with-sm3"),
		strings.Contains(lower, "sm2signwithsm3"),
		strings.Contains(lower, "sm3-with-sm2"),
		strings.Contains(lower, "sm2signingsm3"):
		return "SM2SigningwithSM3"

	// ---- ML-DSA ----
	case strings.Contains(lower, "ml-dsa-44"),
		strings.Contains(lower, "mldsa44"),
		strings.Contains(lower, "dilithium2"):
		return "ML-DSA-44"
	case strings.Contains(lower, "ml-dsa-65"),
		strings.Contains(lower, "mldsa65"),
		strings.Contains(lower, "dilithium3"):
		return "ML-DSA-65"
	case strings.Contains(lower, "ml-dsa-87"),
		strings.Contains(lower, "mldsa87"),
		strings.Contains(lower, "dilithium5"):
		return "ML-DSA-87"

	// ---- Falcon ----
	case strings.Contains(lower, "falcon(512)"),
		strings.Contains(lower, "falcon-512"),
		strings.Contains(lower, "falcon512"):
		return "Falcon(512)"
	case strings.Contains(lower, "falcon(1024)"),
		strings.Contains(lower, "falcon-1024"),
		strings.Contains(lower, "falcon1024"):
		return "Falcon(1024)"

	// ---- SLH-DSA ----
	case strings.Contains(lower, "slh-dsa-sha2-128s"):
		return "SLH-DSA-SHA2-128s"
	case strings.Contains(lower, "slh-dsa-sha2-128f"):
		return "SLH-DSA-SHA2-128f"
	case strings.Contains(lower, "slh-dsa-sha2-192s"):
		return "SLH-DSA-SHA2-192s"
	case strings.Contains(lower, "slh-dsa-sha2-192f"):
		return "SLH-DSA-SHA2-192f"
	case strings.Contains(lower, "slh-dsa-sha2-256s"):
		return "SLH-DSA-SHA2-256s"
	case strings.Contains(lower, "slh-dsa-sha2-256f"):
		return "SLH-DSA-SHA2-256f"
	case strings.Contains(lower, "slh-dsa-shake-128s"):
		return "SLH-DSA-SHAKE-128s"
	case strings.Contains(lower, "slh-dsa-shake-128f"):
		return "SLH-DSA-SHAKE-128f"
	case strings.Contains(lower, "slh-dsa-shake-192s"):
		return "SLH-DSA-SHAKE-192s"
	case strings.Contains(lower, "slh-dsa-shake-192f"):
		return "SLH-DSA-SHAKE-192f"
	case strings.Contains(lower, "slh-dsa-shake-256s"):
		return "SLH-DSA-SHAKE-256s"
	case strings.Contains(lower, "slh-dsa-shake-256f"):
		return "SLH-DSA-SHAKE-256f"

	// ---- RSA 系列 ----
	case strings.Contains(lower, "sha256withrsaencryption"),
		strings.Contains(lower, "sha256withrsa"):
		return "SHA256withRSA"
	case strings.Contains(lower, "sha384withrsaencryption"),
		strings.Contains(lower, "sha384withrsa"):
		return "SHA384withRSA"
	case strings.Contains(lower, "sha512withrsaencryption"),
		strings.Contains(lower, "sha512withrsa"):
		return "SHA512withRSA"
	case strings.Contains(lower, "rsassa-pss"),
		strings.Contains(lower, "rsa-pss"):
		return "RSASSA-PSS"

	// ---- ECDSA 系列 ----
	case strings.Contains(lower, "ecdsa-with-sha256"):
		return "ECDSAWithSHA256"
	case strings.Contains(lower, "ecdsa-with-sha384"):
		return "ECDSAWithSHA384"
	case strings.Contains(lower, "ecdsa-with-sha512"):
		return "ECDSAWithSHA512"
	}

	// 裸 OID 兜底
	if isPureOID(s) {
		return "OID-" + s
	}
	return s
}

// isPureOID 判断字符串是否为纯 OID 形式（如 "1.2.3.4" / "1.3.9999.3.11"）。
//
// 判定规则：只包含数字和 '.'，且至少有一个 '.'，且不以 '.' 开头/结尾。
func isPureOID(s string) bool {
	if s == "" || strings.HasPrefix(s, ".") || strings.HasSuffix(s, ".") {
		return false
	}
	if !strings.Contains(s, ".") {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || c == '.') {
			return false
		}
	}
	return true
}

// =============================================================================
// openssl 完整解析
// =============================================================================

func (p *CertParser) parseWithOpenSSL(bin, certPath string) (*CertDetail, error) {
	textOut, err := RunOpenSSL(bin, "x509", "-in", certPath, "-noout", "-text", "-nameopt", "RFC2253")
	if err != nil {
		return nil, err
	}
	detail := &CertDetail{}
	parseCertText(textOut, detail)

	if fpOut, err := RunOpenSSL(bin, "x509", "-in", certPath, "-noout", "-fingerprint", "-sha1"); err == nil {
		fp := strings.TrimSpace(fpOut)
		if i := strings.Index(fp, "="); i >= 0 {
			fp = strings.TrimSpace(fp[i+1:])
		}
		detail.Fingerprint = strings.ToLower(cleanHex(fp))
	}
	return detail, nil
}

// extractSPKIHex 提取证书公钥的完整 SubjectPublicKeyInfo DER，返回带空格的 hex。
func (p *CertParser) extractSPKIHex(opensslBin, certAbs string) (string, error) {
	pubPEM, err := RunOpenSSL(opensslBin, "x509", "-in", certAbs, "-pubkey", "-noout")
	if err != nil {
		return "", fmt.Errorf("openssl x509 -pubkey: %w", err)
	}
	tmpDir, err := os.MkdirTemp("", "spki-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	pubPath := filepath.Join(tmpDir, "pub.pem")
	derPath := filepath.Join(tmpDir, "pub.der")
	if err := os.WriteFile(pubPath, []byte(pubPEM), 0600); err != nil {
		return "", err
	}
	_, stderr, err := RunOpenSSLFull(opensslBin, "pkey",
		"-pubin", "-in", pubPath, "-outform", "DER", "-out", derPath)
	if err != nil {
		return "", fmt.Errorf("openssl pkey -outform DER: %s", strings.TrimSpace(stderr))
	}
	der, err := os.ReadFile(derPath)
	if err != nil {
		return "", err
	}
	return formatHexBytes(der), nil
}

// formatHexBytes 把字节序列格式化为 "30 59 30 13 ..." 形式。
func formatHexBytes(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	parts := make([]string, len(b))
	for i, v := range b {
		parts[i] = fmt.Sprintf("%02x", v)
	}
	return strings.Join(parts, " ")
}

// =============================================================================
// Go 标准库解析（补充 SPKI / 签名 DER）
// =============================================================================

func (p *CertParser) parseFromPEM(certAbs string) (*CertDetail, error) {
	certPEM, err := os.ReadFile(certAbs)
	if err != nil {
		return nil, exception.New(exception.CodeNotFound, "证书文件不存在", 404, nil)
	}
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, exception.New(exception.CodeParamInvalid, "PEM 解析失败", 400, nil)
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, exception.New(
			exception.CodeParamInvalid,
			fmt.Sprintf("证书解析失败（可能是不支持的算法）: %v", err), 400, nil,
		)
	}

	fp := sha1.Sum(cert.Raw)
	return &CertDetail{
		Version:            fmt.Sprintf("v%d", cert.Version),
		Serial:             strings.ToLower(cert.SerialNumber.Text(16)),
		Issuer:             cert.Issuer.String(),
		Subject:            cert.Subject.String(),
		Fingerprint:        strings.ToLower(hex.EncodeToString(fp[:])),
		NotBefore:          cert.NotBefore.In(cstZone).Format("2006-01-02 15:04:05"),
		NotAfter:           cert.NotAfter.In(cstZone).Format("2006-01-02 15:04:05"),
		PublicKeyAlgorithm: cert.PublicKeyAlgorithm.String(),
		SignatureAlgorithm: cert.SignatureAlgorithm.String(),
		SignatureValue:     hex.EncodeToString(cert.Signature),
		PublicKeyValue:     hex.EncodeToString(cert.RawSubjectPublicKeyInfo),
	}, nil
}

// =============================================================================
// DN 顺序归一化
// =============================================================================

// normalizeDNOrder 把 DN 字符串按照 C → ST → L → O → OU → CN → 其他 的顺序重排。
func normalizeDNOrder(dn string) string {
	dn = strings.TrimSpace(dn)
	if dn == "" {
		return ""
	}
	parts := splitDN(dn)
	order := map[string]int{
		"C": 1, "ST": 2, "L": 3, "O": 4, "OU": 5, "CN": 6,
	}
	type kv struct {
		k, v string
		ord  int
	}
	var items []kv
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		idx := strings.Index(part, "=")
		if idx <= 0 {
			continue
		}
		k := strings.ToUpper(strings.TrimSpace(part[:idx]))
		v := strings.TrimSpace(part[idx+1:])
		ord := 99
		if o, ok := order[k]; ok {
			ord = o
		}
		items = append(items, kv{k, v, ord})
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].ord < items[j].ord
	})
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.k+"="+it.v)
	}
	return strings.Join(out, ",")
}

// =============================================================================
// openssl 输出文本解析辅助
// =============================================================================

// RunOpenSSLFull 执行 openssl 命令。
func RunOpenSSLFull(bin string, args ...string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	cmd.Stdin = bytes.NewReader(nil)
	err := cmd.Run()
	return outBuf.String(), errBuf.String(), err
}

// RunOpenSSL 执行 openssl 命令并返回 stdout。
func RunOpenSSL(bin string, args ...string) (string, error) {
	stdout, _, err := RunOpenSSLFull(bin, args...)
	return stdout, err
}

// extractExtLines 从 openssl x509 -ext <name> 输出中提取扩展值。
func extractExtLines(out string) []string {
	var lines []string
	for _, raw := range strings.Split(out, "\n") {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "X509v3 ") {
			idx := strings.Index(trimmed, ":")
			if idx >= 0 {
				rest := strings.TrimSpace(trimmed[idx+1:])
				rest = strings.TrimSpace(strings.TrimPrefix(rest, "critical"))
				if rest != "" {
					lines = append(lines, rest)
				}
			}
			continue
		}
		lines = append(lines, trimmed)
	}
	return lines
}

// parseCertText 从 openssl x509 -text 输出解析证书字段。
func parseCertText(text string, detail *CertDetail) {
	lines := strings.Split(text, "\n")
	var (
		sigAlgCount   int
		inPubKey      bool
		inSigValue    bool
		inSerial      bool
		inKeyUsage    bool
		inEKU         bool
		pubKeyLines   []string
		sigValueLines []string
		serialLines   []string
		kuLines       []string
		ekuLines      []string
		pubAlg        string
	)
	for _, raw := range lines {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}
		if inSigValue {
			if isHexLine(trimmed) {
				sigValueLines = append(sigValueLines, cleanHex(trimmed))
			}
			continue
		}
		if inKeyUsage {
			if strings.HasPrefix(trimmed, "X509v3 ") ||
				strings.HasPrefix(trimmed, "Signature Algorithm:") {
				inKeyUsage = false
			} else {
				kuLines = append(kuLines, trimmed)
				continue
			}
		}
		if inEKU {
			if strings.HasPrefix(trimmed, "X509v3 ") ||
				strings.HasPrefix(trimmed, "Signature Algorithm:") {
				inEKU = false
			} else {
				ekuLines = append(ekuLines, trimmed)
				continue
			}
		}
		if inPubKey {
			if isHexLine(trimmed) {
				pubKeyLines = append(pubKeyLines, cleanHex(trimmed))
				continue
			}
			if strings.HasPrefix(trimmed, "ASN1 OID") ||
				strings.HasPrefix(trimmed, "NIST CURVE") ||
				strings.HasPrefix(trimmed, "pub:") ||
				strings.HasPrefix(trimmed, "Public-Key:") {
				continue
			}
			inPubKey = false
		}
		if inSerial {
			if isHexLine(trimmed) {
				serialLines = append(serialLines, cleanHex(trimmed))
				continue
			}
			inSerial = false
			if detail.Serial == "" {
				detail.Serial = strings.Join(serialLines, "")
			}
		}
		switch {
		case strings.HasPrefix(trimmed, "Version:"):
			v := strings.TrimSpace(strings.TrimPrefix(trimmed, "Version:"))
			if idx := strings.Index(v, " "); idx > 0 {
				v = v[:idx]
			}
			if v != "" {
				detail.Version = "v" + v
			}
		case strings.HasPrefix(trimmed, "Serial Number:"):
			rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "Serial Number:"))
			if rest != "" && isHexLine(rest) {
				serialLines = append(serialLines, cleanHex(rest))
			}
			inSerial = true
		case strings.HasPrefix(trimmed, "Signature Algorithm:"):
			sigAlgCount++
			if sigAlgCount == 1 && detail.SignatureAlgorithm == "" {
				detail.SignatureAlgorithm = strings.TrimSpace(strings.TrimPrefix(trimmed, "Signature Algorithm:"))
			} else if sigAlgCount >= 2 {
				inSigValue = true
			}
		case strings.HasPrefix(trimmed, "Issuer:"):
			detail.Issuer = strings.TrimSpace(strings.TrimPrefix(trimmed, "Issuer:"))
		case strings.HasPrefix(trimmed, "Subject:"):
			detail.Subject = strings.TrimSpace(strings.TrimPrefix(trimmed, "Subject:"))
		case strings.HasPrefix(trimmed, "Not Before:"):
			detail.NotBefore = normalizeTime(strings.TrimSpace(strings.TrimPrefix(trimmed, "Not Before:")))
		case strings.HasPrefix(trimmed, "Not After"):
			v := strings.TrimSpace(strings.TrimPrefix(trimmed, "Not After"))
			v = strings.TrimSpace(strings.TrimPrefix(v, ":"))
			detail.NotAfter = normalizeTime(v)
		case strings.HasPrefix(trimmed, "Public Key Algorithm:"):
			pubAlg = strings.TrimSpace(strings.TrimPrefix(trimmed, "Public Key Algorithm:"))
		case strings.HasPrefix(trimmed, "ASN1 OID:"):
			oid := strings.TrimSpace(strings.TrimPrefix(trimmed, "ASN1 OID:"))
			if oid != "" && (pubAlg == "id-ecPublicKey" || pubAlg == "") {
				pubAlg = oid
			}
		case strings.HasPrefix(trimmed, "pub:"):
			inPubKey = true
		case strings.HasPrefix(trimmed, "Public-Key:"):
			inPubKey = true
			rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "Public-Key:"))
			if rest != "" && isHexLine(rest) {
				pubKeyLines = append(pubKeyLines, cleanHex(rest))
			}
		case strings.HasPrefix(trimmed, "SHA256 Fingerprint="):
			detail.Fingerprint = strings.ToLower(cleanHex(strings.TrimSpace(strings.TrimPrefix(trimmed, "SHA256 Fingerprint="))))
		case strings.HasPrefix(trimmed, "SHA1 Fingerprint="):
			detail.Fingerprint = strings.ToLower(cleanHex(strings.TrimSpace(strings.TrimPrefix(trimmed, "SHA1 Fingerprint="))))
		case strings.HasPrefix(trimmed, "X509v3 Key Usage:"):
			inKeyUsage = true
			rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "X509v3 Key Usage:"))
			rest = strings.TrimSpace(strings.TrimPrefix(rest, "critical"))
			if rest != "" {
				kuLines = append(kuLines, rest)
			}
		case strings.HasPrefix(trimmed, "X509v3 Extended Key Usage:"):
			inEKU = true
			rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "X509v3 Extended Key Usage:"))
			if rest != "" {
				ekuLines = append(ekuLines, rest)
			}
		}
	}
	if inSerial && detail.Serial == "" {
		detail.Serial = strings.Join(serialLines, "")
	}
	if detail.Serial == "" && len(serialLines) > 0 {
		detail.Serial = strings.Join(serialLines, "")
	}
	if detail.Serial != "" {
		detail.Serial = strings.ToLower(detail.Serial)
	}
	if detail.PublicKeyAlgorithm == "" {
		detail.PublicKeyAlgorithm = pubAlg
	}
	if len(pubKeyLines) > 0 && detail.PublicKeyValue == "" {
		detail.PublicKeyValue = strings.Join(pubKeyLines, "")
	}
	if len(sigValueLines) > 0 && detail.SignatureValue == "" {
		detail.SignatureValue = strings.Join(sigValueLines, "")
	}
	if len(kuLines) > 0 {
		detail.KeyUsage = strings.Join(kuLines, ", ")
	}
	if len(ekuLines) > 0 {
		detail.ExtendedKeyUsage = strings.Join(ekuLines, ", ")
	}
}

// isHexLine 判断是否由十六进制字符与分隔符（冒号/空格/制表符）组成。
func isHexLine(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	s = strings.TrimSuffix(s, ":")
	if s == "" {
		return false
	}
	for _, c := range s {
		if c == ':' || c == ' ' || c == '\t' {
			continue
		}
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func cleanHex(s string) string {
	s = strings.ReplaceAll(s, ":", "")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\t", "")
	return strings.TrimSpace(s)
}

// normalizeTime 把 openssl 输出的时间统一格式化为北京时间（UTC+8）。
func normalizeTime(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		"Jan _2 15:04:05 2006 GMT",
		"Jan 2 15:04:05 2006 GMT",
		"Jan _2 15:04:05 2006 MST",
		"Jan 2 15:04:05 2006 MST",
		"2006-01-02 15:04:05 MST",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.In(cstZone).Format("2006-01-02 15:04:05")
		}
	}
	return s
}

// =============================================================================
// 未使用但保留：避免 import 悬空
// =============================================================================

var _ = pkix.AlgorithmIdentifier{}
var _ = asn1.ObjectIdentifier{}
