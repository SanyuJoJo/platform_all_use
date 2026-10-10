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

// Parse 解析证书。
//
// ★ 解析策略（同时用 openssl + Go 标准库，取长补短）：
//
//  1. openssl 优先：完整解析所有字段（serial/subject/issuer/时间/版本/算法/KeyUsage/EKU），
//     能正确处理 SM2 等 Go 标准库不支持的算法；
//  2. Go 标准库补充：如果 Go 能成功解析（RSA/ECDSA/Ed25519），用 Go 的
//     RawSubjectPublicKeyInfo（完整 SPKI DER）与 Signature（签名 DER）
//     覆盖 openssl 输出的公钥点与签名截断值；
//  3. 展示名归一化：把 openssl 的 OID 原始名映射为第三方平台展示名
//     （id-ecPublicKey → ECC(256)，SM2-with-SM3 → SM2SigningwithSM3）；
//  4. ★ 与第三方平台对齐：摘要值 SHA1、时间 UTC+8、DN 顺序 C→ST→L→O→OU→CN、
//     公钥值完整 SPKI DER hex（带空格）、签名值完整连续小写 hex。
func (p *CertParser) Parse(certAbs string) (*CertDetail, error) {
	// 1. openssl 完整解析
	var detail *CertDetail
	if bin := p.OpensslBin(); bin != "" {
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

	// 3. ★ 统一展示名
	detail.PublicKeyAlgorithm = normalizePublicKeyAlg(detail.PublicKeyAlgorithm)
	detail.SignatureAlgorithm = normalizeSignatureAlg(detail.SignatureAlgorithm)

	// 4. 补全 openssl 解析的兜底字段
	if detail.Serial == "" {
		if bin := p.OpensslBin(); bin != "" {
			if serialOut, err := RunOpenSSL(bin, "x509", "-in", certAbs, "-noout", "-serial"); err == nil {
				s := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(serialOut), "serial="))
				detail.Serial = strings.ToLower(cleanHex(s))
			}
		}
	}
	if detail.KeyUsage == "" {
		if bin := p.OpensslBin(); bin != "" {
			if kuOut, err := RunOpenSSL(bin, "x509", "-in", certAbs, "-noout", "-ext", "keyUsage"); err == nil {
				lines := extractExtLines(kuOut)
				if len(lines) > 0 {
					detail.KeyUsage = strings.Join(lines, ", ")
				}
			}
		}
	}
	if detail.ExtendedKeyUsage == "" {
		if bin := p.OpensslBin(); bin != "" {
			if ekuOut, err := RunOpenSSL(bin, "x509", "-in", certAbs, "-noout", "-ext", "extendedKeyUsage"); err == nil {
				lines := extractExtLines(ekuOut)
				if len(lines) > 0 {
					detail.ExtendedKeyUsage = strings.Join(lines, ", ")
				}
			}
		}
	}

	// 5. ★ DN 顺序统一为 C → ST → L → O → OU → CN
	detail.Issuer = normalizeDNOrder(detail.Issuer)
	detail.Subject = normalizeDNOrder(detail.Subject)

	// 6. ★ 公钥值统一为完整 SPKI DER hex（字节间空格分隔）
	//    以 openssl 输出为准，覆盖 Go 标准库（因 Go 无法解析 SM2）。
	if bin := p.OpensslBin(); bin != "" {
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
// openssl 完整解析
// =============================================================================

func (p *CertParser) parseWithOpenSSL(bin, certPath string) (*CertDetail, error) {
	textOut, err := RunOpenSSL(bin, "x509", "-in", certPath, "-noout", "-text", "-nameopt", "RFC2253")
	if err != nil {
		return nil, err
	}
	detail := &CertDetail{}
	parseCertText(textOut, detail)

	// ★ 摘要值：SHA1（与第三方平台对齐），小写 hex
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
//
// 步骤：
//   openssl x509 -pubkey -noout        → 公钥 PEM
//   openssl pkey -pubin -outform DER   → SPKI DER
//   每个字节格式化为 %02x，空格分隔。
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

// parseFromPEM 用 Go 标准库解析证书。
//
// 用途：提取 RawSubjectPublicKeyInfo（完整 SPKI DER）与 Signature（签名 DER）。
// 注意：Go 标准库不支持 SM2，SM2 证书会返回 error。
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

	// ★ 摘要值：SHA1（与第三方平台对齐），小写 hex
	fp := sha1.Sum(cert.Raw)

	// 公钥值：SPKI DER hex（小写，无分隔符）—— 仅作兜底，会被 openssl 覆盖
	pubKeyValue := hex.EncodeToString(cert.RawSubjectPublicKeyInfo)
	// 签名值：签名 DER hex（小写，无分隔符）
	signValue := hex.EncodeToString(cert.Signature)

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
		SignatureValue:     signValue,
		PublicKeyValue:     pubKeyValue,
	}, nil
}

// =============================================================================
// 展示名归一化
// =============================================================================

// normalizePublicKeyAlg 归一化公钥算法展示名。
func normalizePublicKeyAlg(alg string) string {
	s := strings.TrimSpace(alg)
	if s == "" {
		return ""
	}
	lower := strings.ToLower(s)

	switch {
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
	case strings.Contains(lower, "rsaencryption"),
		strings.Contains(lower, "rsa"):
		return "RSA"
	case strings.Contains(lower, "ed25519"):
		return "Ed25519"
	case strings.Contains(lower, "ed448"):
		return "Ed448"
	case strings.Contains(lower, "dsa"):
		return "DSA"
	}
	return s
}

// normalizeSignatureAlg 归一化签名算法展示名。
func normalizeSignatureAlg(alg string) string {
	s := strings.TrimSpace(alg)
	if s == "" {
		return ""
	}
	lower := strings.ToLower(s)

	switch {
	case strings.Contains(lower, "sm2-with-sm3"),
		strings.Contains(lower, "sm2signwithsm3"),
		strings.Contains(lower, "sm3-with-sm2"),
		strings.Contains(lower, "sm2signingsm3"):
		return "SM2SigningwithSM3"
	case strings.Contains(lower, "ecdsa-with-sha256"):
		return "ECDSAWithSHA256"
	case strings.Contains(lower, "ecdsa-with-sha384"):
		return "ECDSAWithSHA384"
	case strings.Contains(lower, "ecdsa-with-sha512"):
		return "ECDSAWithSHA512"
	case strings.Contains(lower, "sha256withrsaencryption"),
		strings.Contains(lower, "sha256withrsa"):
		return "SHA256withRSA"
	case strings.Contains(lower, "sha384withrsaencryption"),
		strings.Contains(lower, "sha384withrsa"):
		return "SHA384withRSA"
	case strings.Contains(lower, "sha512withrsaencryption"),
		strings.Contains(lower, "sha512withrsa"):
		return "SHA512withRSA"
	}
	return s
}

// =============================================================================
// DN 顺序归一化
// =============================================================================

// normalizeDNOrder 把 DN 字符串按照 C → ST → L → O → OU → CN → 其他 的顺序重排。
//
// 输入形如：
//
//	"CN=213SDA,C=CN"          → "C=CN,CN=213SDA"
//	"CN=AJ2D,C=CN"            → "C=CN,CN=AJ2D"
//	"CN=a,O=b,OU=c,C=CN"      → "C=CN,O=b,OU=c,CN=a"
//
// 说明：
//   - 未知属性的排序值为 99，按原相对顺序保留在最后；
//   - 属性名统一大写；
//   - 值中的转义字符（\, \= 等）原样保留。
func normalizeDNOrder(dn string) string {
	dn = strings.TrimSpace(dn)
	if dn == "" {
		return ""
	}
	parts := splitDN(dn) // 同包 cert_service.go 中定义
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
//
// ★ 修改点：允许空格，兼容 openssl 输出形如 "30 45 02 20 ..." 的签名值。
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
//
// ★ 修改点：由 UTC 改为 UTC+8，与第三方平台对齐。
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
