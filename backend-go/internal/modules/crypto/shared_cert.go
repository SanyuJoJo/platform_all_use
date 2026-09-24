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

// CADetail 证书解析结果。
type CADetail struct {
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
}

func (p *CertParser) Parse(certAbs string) (*CADetail, error) {
	if bin := p.OpensslBin(); bin != "" {
		if detail, err := p.parseWithOpenSSL(bin, certAbs); err == nil && detail.Subject != "" {
			return detail, nil
		}
	}
	return p.parseFromPEM(certAbs)
}

func (p *CertParser) OpensslBin() string {
	candidates := []string{
		filepath.Join(p.coreRoot, "run/bin/openssl"),
		filepath.Join(p.coreRoot, "libs/bin/tongsuo/bin/openssl"),
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && !st.IsDir() && st.Mode()&0111 != 0 {
			return c
		}
	}
	if p, err := exec.LookPath("openssl"); err == nil {
		return p
	}
	return ""
}

func (p *CertParser) parseWithOpenSSL(bin, certPath string) (*CADetail, error) {
	textOut, err := RunOpenSSL(bin, "x509", "-in", certPath, "-noout", "-text", "-nameopt", "RFC2253")
	if err != nil {
		return nil, err
	}
	detail := &CADetail{}
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

func (p *CertParser) parseFromPEM(certAbs string) (*CADetail, error) {
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
	return &CADetail{
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
	}, nil
}

// -----------------------------------------------------------------------------
// OpenSSL 命令执行与文本解析辅助
// -----------------------------------------------------------------------------

// RunOpenSSLFull 执行 openssl 命令，返回 stdout / stderr / error。
func RunOpenSSLFull(bin string, args ...string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	return outBuf.String(), errBuf.String(), err
}

// RunOpenSSL 执行 openssl 命令并返回 stdout。
func RunOpenSSL(bin string, args ...string) (string, error) {
	stdout, _, err := RunOpenSSLFull(bin, args...)
	return stdout, err
}

func parseCertText(text string, detail *CADetail) {
	lines := strings.Split(text, "\n")
	var (
		sigAlgCount   int
		inPubKey      bool
		inSigValue    bool
		inSerial      bool
		pubKeyLines   []string
		sigValueLines []string
		serialLines   []string
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
