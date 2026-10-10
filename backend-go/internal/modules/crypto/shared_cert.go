package crypto

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

// CADetail 是 CertDetail 的类型别名。
type CADetail = CertDetail

// Parse 解析证书。
func (p *CertParser) Parse(certAbs string) (*CertDetail, error) {
	if bin := p.OpensslBin(); bin != "" {
		if detail, err := p.parseWithOpenSSL(bin, certAbs); err == nil && detail.Subject != "" {
			if detail.Serial == "" {
				if serialOut, err := RunOpenSSL(bin, "x509", "-in", certAbs, "-noout", "-serial"); err == nil {
					serialOut = strings.TrimSpace(serialOut)
					serialOut = strings.TrimPrefix(serialOut, "serial=")
					detail.Serial = cleanHex(serialOut)
				}
			}
			if detail.KeyUsage == "" {
				if kuOut, err := RunOpenSSL(bin, "x509", "-in", certAbs, "-noout", "-ext", "keyUsage"); err == nil {
					lines := extractExtLines(kuOut)
					if len(lines) > 0 {
						detail.KeyUsage = strings.Join(lines, ", ")
					}
				}
			}
			if detail.ExtendedKeyUsage == "" {
				if ekuOut, err := RunOpenSSL(bin, "x509", "-in", certAbs, "-noout", "-ext", "extendedKeyUsage"); err == nil {
					lines := extractExtLines(ekuOut)
					if len(lines) > 0 {
						detail.ExtendedKeyUsage = strings.Join(lines, ", ")
					}
				}
			}
			return detail, nil
		}
	}
	return p.parseFromPEM(certAbs)
}

// OpensslBin 返回铜锁 openssl 路径。
//
// ★ 候选路径（按优先级）：
//  1. <coreRoot>/libs/tongsuo/bin/openssl      （部署后新规范）
//  2. <coreRoot>/libs/bin/tongsuo/bin/openssl  （历史布局）
//  3. <coreRoot>/run/bin/openssl               （最旧布局）
//
// 全部不存在时才回退系统 openssl。
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

func (p *CertParser) parseWithOpenSSL(bin, certPath string) (*CertDetail, error) {
	textOut, err := RunOpenSSL(bin, "x509", "-in", certPath, "-noout", "-text", "-nameopt", "RFC2253")
	if err != nil {
		return nil, err
	}
	detail := &CertDetail{}
	parseCertText(textOut, detail)
	if fpOut, err := RunOpenSSL(bin, "x509", "-in", certPath, "-noout", "-fingerprint", "-sha256"); err == nil {
		fp := strings.TrimSpace(fpOut)
		if i := strings.Index(fp, "="); i >= 0 {
			fp = strings.TrimSpace(fp[i+1:])
		}
		detail.Fingerprint = cleanHex(fp)
	}
	return detail, nil
}

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
	fp := sha256.Sum256(cert.Raw)

	kuParts := []string{}
	if cert.KeyUsage&x509.KeyUsageDigitalSignature != 0 {
		kuParts = append(kuParts, "Digital Signature")
	}
	if cert.KeyUsage&x509.KeyUsageContentCommitment != 0 {
		kuParts = append(kuParts, "Non Repudiation")
	}
	if cert.KeyUsage&x509.KeyUsageKeyEncipherment != 0 {
		kuParts = append(kuParts, "Key Encipherment")
	}
	if cert.KeyUsage&x509.KeyUsageDataEncipherment != 0 {
		kuParts = append(kuParts, "Data Encipherment")
	}
	if cert.KeyUsage&x509.KeyUsageKeyAgreement != 0 {
		kuParts = append(kuParts, "Key Agreement")
	}
	if cert.KeyUsage&x509.KeyUsageCertSign != 0 {
		kuParts = append(kuParts, "Certificate Sign")
	}
	if cert.KeyUsage&x509.KeyUsageCRLSign != 0 {
		kuParts = append(kuParts, "CRL Sign")
	}

	ekuParts := []string{}
	for _, e := range cert.ExtKeyUsage {
		switch e {
		case x509.ExtKeyUsageServerAuth:
			ekuParts = append(ekuParts, "TLS Web Server Authentication")
		case x509.ExtKeyUsageClientAuth:
			ekuParts = append(ekuParts, "TLS Web Client Authentication")
		case x509.ExtKeyUsageCodeSigning:
			ekuParts = append(ekuParts, "Code Signing")
		case x509.ExtKeyUsageEmailProtection:
			ekuParts = append(ekuParts, "E-mail Protection")
		case x509.ExtKeyUsageTimeStamping:
			ekuParts = append(ekuParts, "Time Stamping")
		case x509.ExtKeyUsageOCSPSigning:
			ekuParts = append(ekuParts, "OCSP Signing")
		}
	}

	return &CertDetail{
		Version:            fmt.Sprintf("v%d", cert.Version),
		Serial:             cert.SerialNumber.Text(16),
		Issuer:             cert.Issuer.String(),
		Subject:            cert.Subject.String(),
		Fingerprint:        hex.EncodeToString(fp[:]),
		NotBefore:          cert.NotBefore.UTC().Format("2006-01-02 15:04:05"),
		NotAfter:           cert.NotAfter.UTC().Format("2006-01-02 15:04:05"),
		PublicKeyAlgorithm: cert.PublicKeyAlgorithm.String(),
		SignatureAlgorithm: cert.SignatureAlgorithm.String(),
		SignatureValue:     hex.EncodeToString(cert.Signature),
		PublicKeyValue:     hex.EncodeToString(cert.RawSubjectPublicKeyInfo),
		KeyUsage:           strings.Join(kuParts, ", "),
		ExtendedKeyUsage:   strings.Join(ekuParts, ", "),
	}, nil
}

// -----------------------------------------------------------------------------
// OpenSSL 命令执行与文本解析辅助
// -----------------------------------------------------------------------------

// RunOpenSSLFull 执行 openssl 命令。
//
// ★ 关键：显式设置 cmd.Stdin = bytes.NewReader(nil)，禁止 openssl 从 stdin 读取；
//   同时将所有 openssl 调用加 -passin pass: 之类的显式密码来源（由调用方决定）。
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
			detail.Fingerprint = cleanHex(strings.TrimSpace(strings.TrimPrefix(trimmed, "SHA256 Fingerprint=")))
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
	if detail.PublicKeyAlgorithm == "" {
		detail.PublicKeyAlgorithm = pubAlg
	}
	if len(pubKeyLines) > 0 {
		detail.PublicKeyValue = strings.Join(pubKeyLines, "")
	}
	if len(sigValueLines) > 0 {
		detail.SignatureValue = strings.Join(sigValueLines, "")
	}
	if len(kuLines) > 0 {
		detail.KeyUsage = strings.Join(kuLines, ", ")
	}
	if len(ekuLines) > 0 {
		detail.ExtendedKeyUsage = strings.Join(ekuLines, ", ")
	}
}

func isHexLine(s string) bool {
	if len(s) < 2 {
		return false
	}
	s = strings.TrimSuffix(s, ":")
	if s == "" {
		return false
	}
	for _, c := range s {
		if c == ':' {
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
			return t.UTC().Format("2006-01-02 15:04:05")
		}
	}
	return s
}
