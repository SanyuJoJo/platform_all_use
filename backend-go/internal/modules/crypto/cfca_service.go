package crypto

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"backend-go/internal/config"
	"backend-go/internal/exception"
	"backend-go/internal/models"
)

// defaultCFCATmpDir 是内置固定临时密钥对的默认目录。
//
// 可通过环境变量 CFCA_TMP_KEY_DIR 覆盖。
//
// 目录内容：
//
//	cfca_tmp.key   0600，SM2 私钥 PEM（固定）
//	cfca_tmp.pub   0644，SM2 公钥 PEM（固定）
//
// 首次启动时若缺失会自动生成；之后所有 CFCA P10 生成 / 双证签发 / 信封解密
// 均复用同一对密钥。
const defaultCFCATmpDir = "/opt/plats_tool/user_data/cert/cfca_tmp"

func resolveCFCATmpDir() string {
	if v := strings.TrimSpace(os.Getenv("CFCA_TMP_KEY_DIR")); v != "" {
		return v
	}
	return defaultCFCATmpDir
}

func oidTmpPubString() string {
	return OIDCFCATmpPublicKey.String()
}

type CFCAService struct {
	db          *gorm.DB
	caller      *CoreCaller
	files       *FileStore
	parser      *CertParser
	keys        *KeyCrypto
	layout      *PathLayout
	whitebox    *WhiteboxClient
	caSvc       *CAService
	envelopeBin string

	// ★ 内置固定临时密钥对
	tmpKeyDir  string
	tmpKeyPath string
	tmpPubPath string
}

func NewCFCAService(
	db *gorm.DB, caller *CoreCaller, files *FileStore,
	parser *CertParser, keys *KeyCrypto,
	layout *PathLayout, whitebox *WhiteboxClient,
	caSvc *CAService, cfg *config.Config,
) *CFCAService {
	binDir := ""
	if s := strings.TrimSpace(cfg.CoreGoBin); s != "" {
		binDir = filepath.Dir(s)
	} else if s := strings.TrimSpace(cfg.CoreDispatchPath); s != "" {
		binDir = filepath.Dir(s)
	}
	tmpDir := resolveCFCATmpDir()
	return &CFCAService{
		db: db, caller: caller, files: files, parser: parser, keys: keys,
		layout: layout, whitebox: whitebox, caSvc: caSvc,
		envelopeBin: filepath.Join(binDir, "cfca_envelope"),
		tmpKeyDir:   tmpDir,
		tmpKeyPath:  filepath.Join(tmpDir, "cfca_tmp.key"),
		tmpPubPath:  filepath.Join(tmpDir, "cfca_tmp.pub"),
	}
}

// =============================================================================
// ★ 内置固定临时密钥对管理
// =============================================================================

// ensureTmpKeyPair 确保内置固定临时密钥对存在（幂等）。
func (s *CFCAService) ensureTmpKeyPair() error {
	if err := os.MkdirAll(s.tmpKeyDir, 0700); err != nil {
		return fmt.Errorf("mkdir %s: %w", s.tmpKeyDir, err)
	}

	keyOK := false
	if st, err := os.Stat(s.tmpKeyPath); err == nil && !st.IsDir() && st.Size() > 0 {
		keyOK = true
	}
	pubOK := false
	if st, err := os.Stat(s.tmpPubPath); err == nil && !st.IsDir() && st.Size() > 0 {
		pubOK = true
	}
	if keyOK && pubOK {
		return nil
	}

	bin := s.parser.OpensslBin()
	if bin == "" {
		return fmt.Errorf("铜锁 openssl 不可用")
	}

	if !keyOK {
		_, stderr, err := RunOpenSSLFull(bin,
			"genpkey", "-algorithm", "SM2", "-out", s.tmpKeyPath)
		if err != nil {
			return fmt.Errorf("生成 CFCA 内置临时私钥失败: %s", strings.TrimSpace(stderr))
		}
		_ = os.Chmod(s.tmpKeyPath, 0600)
		pubOK = false // 私钥变了，公钥必须重导
	}
	if !pubOK {
		_, stderr, err := RunOpenSSLFull(bin,
			"pkey", "-in", s.tmpKeyPath, "-pubout", "-out", s.tmpPubPath)
		if err != nil {
			return fmt.Errorf("导出 CFCA 内置临时公钥失败: %s", strings.TrimSpace(stderr))
		}
		_ = os.Chmod(s.tmpPubPath, 0644)
	}

	log.Info().
		Str("tmp_dir", s.tmpKeyDir).
		Str("tmp_key", s.tmpKeyPath).
		Str("tmp_pub", s.tmpPubPath).
		Msg("cfca: 内置固定临时密钥对已就绪")
	return nil
}

// loadTmpPubDER 读取内置固定临时公钥的 SPKI DER 字节。
func (s *CFCAService) loadTmpPubDER() ([]byte, error) {
	data, err := os.ReadFile(s.tmpPubPath)
	if err != nil {
		return nil, fmt.Errorf("读 CFCA 内置临时公钥失败: %w", err)
	}
	der := pemDER(data)
	if len(der) == 0 {
		return nil, fmt.Errorf("CFCA 内置临时公钥 PEM 解析为空")
	}
	return der, nil
}

// =============================================================================
// ★ 申请 P10 的 CFCA 分支（由 Handler.Execute 拦截 csr.create 调用）
// =============================================================================

// GenerateCSRForApply 为 "申请 P10" 场景生成 CFCA 格式 CSR。
//
// 与 SignDual 中的 generateBundle 逻辑相同（内置固定临时密钥 → 携带临时公钥
// attribute 的 CSR），但仅返回 PEM 数据，不落盘、不落库。
// 落盘与落库由调用方（CSRService.Import）完成。
func (s *CFCAService) GenerateCSRForApply(
	ctx context.Context, subject map[string]string, san []string,
) (*CFCAGenerateResult, error) {
	if err := s.ensureTmpKeyPair(); err != nil {
		return nil, exception.New(exception.CodeInternalError,
			"确保 CFCA 内置临时密钥失败: "+err.Error(), 500, nil)
	}
	if len(subject) == 0 || strings.TrimSpace(subject["CN"]) == "" {
		return nil, exception.New(exception.CodeParamInvalid,
			"CFCA P10 需要 subject.CN", 400, nil)
	}

	bin := s.parser.OpensslBin()
	if bin == "" {
		return nil, exception.New(exception.CodeInternalError, "铜锁 openssl 不可用", 500, nil)
	}

	req := &CFCASignRequest{
		CSRSource: "generate",
		Subject:   subject,
		SAN:       san,
	}
	resolved, err := s.generateBundle(ctx, bin, req)
	if err != nil {
		return nil, err
	}
	return &CFCAGenerateResult{
		CSRPEM:     resolved.CSRPEM,
		SignKeyPEM: resolved.SignKeyPEM,
		Subject:    subject,
		SAN:        san,
	}, nil
}

func pubkeyFingerprint(bin, pubPEMPath string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "cfca-fp-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)
	derPath := filepath.Join(tmpDir, "pub.der")
	_, stderr, err := RunOpenSSLFull(bin, "pkey",
		"-pubin", "-in", pubPEMPath, "-outform", "DER", "-out", derPath)
	if err != nil {
		return "", fmt.Errorf("pkey -outform DER 失败: %s", strings.TrimSpace(stderr))
	}
	out, err := RunOpenSSL(bin, "dgst", "-sm3", "-r", derPath)
	if err != nil {
		return "", err
	}
	parts := strings.Fields(out)
	if len(parts) < 1 {
		return "", fmt.Errorf("dgst 输出异常: %s", out)
	}
	fp := strings.ToLower(parts[0])
	if len(fp) > 32 {
		fp = fp[:32]
	}
	return fp, nil
}

// =============================================================================
// 主流程
// =============================================================================

func (s *CFCAService) SignDual(
	ctx context.Context, req *CFCASignRequest,
) (*CFCASignResponse, error) {
	if req.ValidityDays <= 0 {
		req.ValidityDays = 365
	}
	if req.CASource == "" {
		req.CASource = "local"
	}
	if req.CSRSource == "" {
		req.CSRSource = "generate"
	}

	if err := s.ensureTmpKeyPair(); err != nil {
		return nil, exception.New(exception.CodeInternalError,
			"确保 CFCA 内置临时密钥失败: "+err.Error(), 500, nil)
	}

	bin := s.parser.OpensslBin()
	if bin == "" {
		return nil, exception.New(exception.CodeInternalError, "铜锁 openssl 不可用", 500, nil)
	}

	resolved, err := s.resolveCSRSource(ctx, req)
	if err != nil {
		return nil, err
	}

	tmpPubDER, err := s.loadTmpPubDER()
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			"读 CFCA 内置临时公钥失败: "+err.Error(), 500, nil)
	}

	caCertPEM, caKeyPEM, err := s.loadCA(ctx, req)
	if err != nil {
		return nil, err
	}

	tmpDir, err := os.MkdirTemp("", "cfca-sign-")
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("创建临时目录失败: %v", err), 500, nil)
	}
	defer os.RemoveAll(tmpDir)

	csrPath := filepath.Join(tmpDir, "sign.csr")
	if err := os.WriteFile(csrPath, resolved.CSRPEM, 0600); err != nil {
		return nil, exception.New(exception.CodeInternalError, "写临时 CSR 失败", 500, nil)
	}

	if csrTmpPubDER, err := s.extractTmpPubDERFromCSR(bin, csrPath); err == nil && len(csrTmpPubDER) > 0 {
		if !bytes.Equal(csrTmpPubDER, tmpPubDER) {
			log.Warn().
				Int("csr_tmp_pub_len", len(csrTmpPubDER)).
				Int("builtin_tmp_pub_len", len(tmpPubDER)).
				Msg("cfca: CSR 中的临时公钥与内置固定公钥不一致（seal 仍使用内置固定公钥，unseal 用内置固定私钥）")
		}
	} else if err != nil {
		log.Warn().Err(err).
			Msg("cfca: 无法从 CSR 提取临时公钥（seal 仍使用内置固定公钥）")
	}

	csrSubject, err := s.extractCSRSubject(bin, csrPath)
	if err != nil {
		log.Warn().Err(err).
			Str("csr_path", csrPath).
			Msg("cfca: CSR subject 提取失败，使用请求中的 subject")
		csrSubject = map[string]string{}
	}
	finalSubject := mergeSubject(req.Subject, csrSubject)
	if strings.TrimSpace(finalSubject["CN"]) == "" {
		return nil, exception.New(exception.CodeParamInvalid,
			"无法确定 subject：请在请求中提供 subject.CN", 400, nil)
	}

	caCertPath := filepath.Join(tmpDir, "ca.crt")
	caKeyPath := filepath.Join(tmpDir, "ca.key")
	if err := os.WriteFile(caCertPath, caCertPEM, 0600); err != nil {
		return nil, exception.New(exception.CodeInternalError, "写临时 CA 证书失败", 500, nil)
	}
	if err := os.WriteFile(caKeyPath, caKeyPEM, 0600); err != nil {
		return nil, exception.New(exception.CodeInternalError, "写临时 CA 私钥失败", 500, nil)
	}

	signCertPath := filepath.Join(tmpDir, "sign.cer")
	if err := s.issueCert(bin, csrPath, caCertPath, caKeyPath, signCertPath,
		req.ValidityDays, req.SAN, false, finalSubject); err != nil {
		return nil, err
	}

	encKeyPath := filepath.Join(tmpDir, "enc.key")
	encCSRPath := filepath.Join(tmpDir, "enc.csr")
	if err := s.genSimpleCSR(bin, encKeyPath, encCSRPath, finalSubject); err != nil {
		return nil, err
	}
	encCertPath := filepath.Join(tmpDir, "enc.cer")
	if err := s.issueCert(bin, encCSRPath, caCertPath, caKeyPath, encCertPath,
		req.ValidityDays, req.SAN, true, finalSubject); err != nil {
		return nil, err
	}

	sealPubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: tmpPubDER})
	sealPubPath := filepath.Join(tmpDir, "seal.tmp.pub")
	if err := os.WriteFile(sealPubPath, sealPubPEM, 0600); err != nil {
		return nil, exception.New(exception.CodeInternalError, "写临时公钥失败", 500, nil)
	}
	if fp, err := pubkeyFingerprint(bin, sealPubPath); err == nil {
		log.Info().Str("seal_tmp_pub_fp", fp).
			Msg("cfca: seal 使用内置固定临时公钥（unseal 用内置固定临时私钥）")
	}

	envelopePath := filepath.Join(tmpDir, "envelope.txt")
	if err := s.sealEnvelope(ctx, sealPubPath, encKeyPath, envelopePath); err != nil {
		return nil, err
	}

	signCertPEM, _ := os.ReadFile(signCertPath)
	encCertPEM, _ := os.ReadFile(encCertPath)
	envelopeBytes, _ := os.ReadFile(envelopePath)
	encKeyPEM, _ := os.ReadFile(encKeyPath)

	persisted, err := s.persist(ctx, req, signCertPEM, encCertPEM, encKeyPEM, resolved.SignKeyPEM)
	if err != nil {
		return nil, err
	}

	resp := &CFCASignResponse{
		SignCertPEM:       string(signCertPEM),
		EncCertPEM:        string(encCertPEM),
		EncryptedEnvelope: string(envelopeBytes),
		SignCertID:        persisted.SignCertID,
		EncCertID:         persisted.EncCertID,
	}
	if req.ReturnSignKey && len(resolved.SignKeyPEM) > 0 {
		resp.SignKeyPEM = string(resolved.SignKeyPEM)
	}
	return resp, nil
}

// =============================================================================
// CSR 来源解析
// =============================================================================

func (s *CFCAService) resolveCSRSource(
	ctx context.Context, req *CFCASignRequest,
) (*cfcaResolvedCSR, error) {
	bin := s.parser.OpensslBin()
	switch req.CSRSource {
	case "generate":
		return s.generateBundle(ctx, bin, req)

	case "upload":
		if strings.TrimSpace(req.CSRPEM) == "" {
			return nil, exception.New(exception.CodeParamInvalid, "upload 模式需要 csr_pem", 400, nil)
		}
		var signKeyPEM []byte
		if strings.TrimSpace(req.CSRKeyPEM) != "" {
			plain, err := s.maybeDecryptKeyPEM(bin, req.CSRKeyPEM, req.CSRKeyPassword)
			if err != nil {
				return nil, err
			}
			signKeyPEM = plain
		}
		return &cfcaResolvedCSR{CSRPEM: []byte(req.CSRPEM), SignKeyPEM: signKeyPEM}, nil

	case "existing":
		if req.CSRID == "" {
			return nil, exception.New(exception.CodeParamInvalid, "existing 模式需要 csr_id", 400, nil)
		}
		var csr models.CSR
		if err := s.db.Where("csr_id = ?", req.CSRID).First(&csr).Error; err != nil {
			return nil, exception.New(exception.CodeNotFound, "CSR 不存在", 404, nil)
		}
		data, _ := os.ReadFile(csr.CSRPath)
		var signKeyPEM []byte
		if csr.KeyRef != nil && *csr.KeyRef != "" {
			if kp, err := s.caller.ExportKey(ctx, *csr.KeyRef); err == nil {
				signKeyPEM = kp
			}
		} else if keyAbs := s.deriveCSRKeyPath(csr); keyAbs != "" {
			if plainTmp, err := s.whitebox.DecryptToTemp(ctx, keyAbs, s.layout.TmpDir); err == nil {
				signKeyPEM, _ = os.ReadFile(plainTmp)
				os.Remove(plainTmp)
			}
		}
		return &cfcaResolvedCSR{CSRPEM: data, SignKeyPEM: signKeyPEM}, nil

	default:
		return nil, exception.New(exception.CodeParamInvalid,
			"csr_source 必须是 generate / upload / existing", 400, nil)
	}
}

// generateBundle 使用内置固定临时私钥生成 CFCA P10。
func (s *CFCAService) generateBundle(
	ctx context.Context, bin string, req *CFCASignRequest,
) (*cfcaResolvedCSR, error) {
	if len(req.Subject) == 0 || strings.TrimSpace(req.Subject["CN"]) == "" {
		return nil, exception.New(exception.CodeParamInvalid,
			"generate 模式需要提供 subject.CN", 400, nil)
	}
	if err := s.ensureTmpKeyPair(); err != nil {
		return nil, exception.New(exception.CodeInternalError,
			"确保 CFCA 内置临时密钥失败: "+err.Error(), 500, nil)
	}

	tmpDir, err := os.MkdirTemp("", "cfca-gen-")
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("创建临时目录失败: %v", err), 500, nil)
	}
	defer os.RemoveAll(tmpDir)

	csrPath := filepath.Join(tmpDir, "sign.csr")
	signKeyPath := filepath.Join(tmpDir, "sign.key")

	subjectStr := buildSubjectString(req.Subject)
	if subjectStr == "" {
		return nil, exception.New(exception.CodeParamInvalid, "subject 不能为空", 400, nil)
	}

	args := []string{
		"gen-csr",
		"--subject", subjectStr,
		"--out-csr", csrPath,
		"--out-sign-key", signKeyPath,
		"--tmp-key", s.tmpKeyPath, // ★ 使用内置固定临时私钥
	}
	cmd := exec.CommandContext(ctx, s.envelopeBin, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdin = nil
	if err := cmd.Run(); err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("cfca_envelope gen-csr 失败: %v, stderr=%s",
				err, strings.TrimSpace(stderr.String())), 500, nil)
	}

	csrPEM, err := os.ReadFile(csrPath)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError, "读 CSR 失败", 500, nil)
	}
	signKeyPEM, err := os.ReadFile(signKeyPath)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError, "读签名私钥失败", 500, nil)
	}

	return &cfcaResolvedCSR{
		CSRPEM:     csrPEM,
		SignKeyPEM: signKeyPEM,
	}, nil
}

// =============================================================================
// csr2pub / csr2subj
// =============================================================================

func (s *CFCAService) extractTmpPubDERFromCSR(_ string, csrPath string) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "cfca-csr2pub-")
	if err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	outPath := filepath.Join(tmpDir, "tmp.pub.der")
	cmd := exec.Command(s.envelopeBin, "csr2pub",
		"--in", csrPath, "--out", outPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdin = nil
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("cfca_envelope csr2pub 失败: %v, stderr=%s",
			err, strings.TrimSpace(stderr.String()))
	}

	der, err := os.ReadFile(outPath)
	if err != nil {
		return nil, fmt.Errorf("读 csr2pub 输出失败: %v", err)
	}
	if len(der) == 0 {
		return nil, fmt.Errorf("csr2pub 输出为空")
	}
	return der, nil
}

func (s *CFCAService) extractCSRSubject(bin, csrPath string) (map[string]string, error) {
	var subj string

	out, err := RunOpenSSL(bin, "req", "-in", csrPath, "-noout", "-subject", "-nameopt", "RFC2253")
	if err == nil {
		subj = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(out), "subject="))
	}

	if strings.TrimSpace(subj) == "" {
		tmpDir, derr := os.MkdirTemp("", "cfca-subj-")
		if derr != nil {
			return nil, fmt.Errorf("创建临时目录失败: %v", derr)
		}
		defer os.RemoveAll(tmpDir)

		outPath := filepath.Join(tmpDir, "subj.txt")
		cmd := exec.Command(s.envelopeBin, "csr2subj",
			"--in", csrPath, "--out", outPath)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		cmd.Stdin = nil
		if rerr := cmd.Run(); rerr != nil {
			return nil, fmt.Errorf(
				"openssl req 与 cfca_envelope csr2subj 均失败；csr2subj stderr=%s",
				strings.TrimSpace(stderr.String()))
		}
		data, rerr := os.ReadFile(outPath)
		if rerr != nil {
			return nil, fmt.Errorf("读 csr2subj 输出失败: %v", rerr)
		}
		subj = strings.TrimSpace(string(data))
	}

	if subj == "" {
		return nil, fmt.Errorf("CSR subject 为空")
	}

	m := map[string]string{}
	for _, part := range strings.Split(subj, ",") {
		idx := strings.Index(part, "=")
		if idx <= 0 {
			continue
		}
		m[strings.ToUpper(strings.TrimSpace(part[:idx]))] = strings.TrimSpace(part[idx+1:])
	}
	return m, nil
}

func mergeSubject(req, csr map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range csr {
		out[k] = v
	}
	for k, v := range req {
		if v != "" {
			out[k] = v
		}
	}
	return out
}

func buildSubjectString(m map[string]string) string {
	var b strings.Builder
	for _, k := range []string{"C", "ST", "L", "O", "OU", "CN", "E"} {
		if v := m[k]; v != "" {
			b.WriteString("/")
			b.WriteString(k)
			b.WriteString("=")
			b.WriteString(v)
		}
	}
	return b.String()
}

// =============================================================================
// openssl 命令封装
// =============================================================================

func (s *CFCAService) genSM2Key(bin, outPath string) error {
	_, stderr, err := RunOpenSSLFull(bin, "genpkey", "-algorithm", "SM2", "-out", outPath)
	if err != nil {
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("生成 SM2 密钥失败: %s", strings.TrimSpace(stderr)), 500, nil)
	}
	return nil
}

func (s *CFCAService) genSimpleCSR(
	_ string, keyPath, csrPath string, subject map[string]string,
) error {
	subj := buildSubjectString(subject)
	if subj == "" {
		return exception.New(exception.CodeParamInvalid, "subject 不能为空", 400, nil)
	}

	cmd := exec.Command(s.envelopeBin, "gen-plain-csr",
		"--subject", subj,
		"--out-csr", csrPath,
		"--out-key", keyPath,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdin = nil
	if err := cmd.Run(); err != nil {
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("生成加密 CSR 失败: %s", strings.TrimSpace(stderr.String())), 500, nil)
	}
	return nil
}

func (s *CFCAService) issueCert(
	_ string, csrPath, caCertPath, caKeyPath, outPath string,
	days int, san []string, isEnc bool, subject map[string]string,
) error {
	args := []string{
		"sign-cert",
		"--csr", csrPath,
		"--ca-cert", caCertPath,
		"--ca-key", caKeyPath,
		"--out-cert", outPath,
		"--days", strconv.Itoa(days),
	}
	if isEnc {
		args = append(args, "--is-enc", "1")
	} else {
		args = append(args, "--is-enc", "0")
	}
	if len(san) > 0 {
		args = append(args, "--san", strings.Join(san, ","))
	}
	if subjStr := buildSubjectString(subject); subjStr != "" {
		args = append(args, "--subject", subjStr)
	}

	cmd := exec.Command(s.envelopeBin, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdin = nil
	if err := cmd.Run(); err != nil {
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("CA 签发证书失败: %s", strings.TrimSpace(stderr.String())), 500, nil)
	}
	return nil
}

func (s *CFCAService) sealEnvelope(ctx context.Context, sealPubPath, encKeyPath, outPath string) error {
	cmd := exec.CommandContext(ctx, s.envelopeBin, "seal",
		"--tmp-pub", sealPubPath,
		"--enc-key", encKeyPath,
		"--out", outPath,
		"--format", "cfca",
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdin = nil
	err := cmd.Run()
	log.Info().
		Str("seal_stderr", strings.TrimSpace(stderr.String())).
		Bool("ok", err == nil).
		Msg("cfca: seal 诊断输出")
	if err != nil {
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("封信封失败: %v, stderr=%s", err, strings.TrimSpace(stderr.String())), 500, nil)
	}
	return nil
}

// =============================================================================
// CA 加载
// =============================================================================

func (s *CFCAService) loadCA(
	ctx context.Context, req *CFCASignRequest,
) (certPEM, keyPEM []byte, err error) {
	switch req.CASource {
	case "local", "":
		if req.CAID == "" {
			return nil, nil, exception.New(exception.CodeParamInvalid, "本地 CA 需要 ca_id", 400, nil)
		}
		ca, e := s.caSvc.Get(req.CAID)
		if e != nil {
			return nil, nil, e
		}
		certPEM, e = s.caSvc.ReadCACert(ca)
		if e != nil {
			return nil, nil, exception.New(exception.CodeInternalError,
				fmt.Sprintf("读 CA 证书失败: %v", e), 500, nil)
		}
		keyPEM, e = s.loadCAPrivKey(ctx, ca)
		if e != nil {
			return nil, nil, e
		}
		return certPEM, keyPEM, nil

	case "manual":
		if strings.TrimSpace(req.CACertPEM) == "" || strings.TrimSpace(req.CAKeyPEM) == "" {
			return nil, nil, exception.New(exception.CodeParamInvalid,
				"manual 模式需要 ca_cert_pem + ca_key_pem", 400, nil)
		}
		kp, e := s.maybeDecryptKeyPEM(s.parser.OpensslBin(), req.CAKeyPEM, req.CAKeyPassword)
		if e != nil {
			return nil, nil, e
		}
		return []byte(req.CACertPEM), kp, nil

	default:
		return nil, nil, exception.New(exception.CodeParamInvalid,
			"ca_source 必须是 local 或 manual", 400, nil)
	}
}

func (s *CFCAService) loadCAPrivKey(ctx context.Context, ca *models.CA) ([]byte, error) {
	if ca.KeyRef != "" {
		return s.caller.ExportKey(ctx, ca.KeyRef)
	}
	if ca.KeyPath != "" {
		keyAbs := ca.KeyPath
		if !filepath.IsAbs(keyAbs) {
			keyAbs = filepath.Join(s.layout.CertRoot, keyAbs)
		}
		plainTmp, err := s.whitebox.DecryptToTemp(ctx, keyAbs, s.layout.TmpDir)
		if err != nil {
			return nil, exception.New(exception.CodeInternalError,
				fmt.Sprintf("解密 CA 白盒私钥失败: %v", err), 500, nil)
		}
		defer os.Remove(plainTmp)
		return os.ReadFile(plainTmp)
	}
	return nil, exception.New(exception.CodeParamInvalid, "该 CA 未关联私钥，无法签发", 400, nil)
}

func (s *CFCAService) maybeDecryptKeyPEM(bin, keyPEM, password string) ([]byte, error) {
	if s.keys == nil {
		return []byte(keyPEM), nil
	}
	tmpDir, _ := os.MkdirTemp("", "cfca-key-")
	defer os.RemoveAll(tmpDir)
	tmpKey := filepath.Join(tmpDir, "k.pem")
	if err := os.WriteFile(tmpKey, []byte(keyPEM), 0600); err != nil {
		return nil, exception.New(exception.CodeInternalError, "写临时私钥失败", 500, nil)
	}
	encrypted, _ := s.keys.IsEncrypted(bin, tmpKey)
	if !encrypted {
		return []byte(keyPEM), nil
	}
	if strings.TrimSpace(password) == "" {
		return nil, exception.New(exception.CodeParamInvalid, "私钥已加密，请提供密码", 400, nil)
	}
	pwPath := filepath.Join(tmpDir, "pw.txt")
	plainPath := filepath.Join(tmpDir, "plain.key")
	os.WriteFile(pwPath, []byte(password), 0600)
	if err := s.keys.DecryptWithPasswordFile(bin, tmpKey, pwPath, plainPath); err != nil {
		return nil, exception.New(exception.CodeParamInvalid, "私钥密码错误或解密失败", 400, nil)
	}
	return os.ReadFile(plainPath)
}

// =============================================================================
// 落盘 + 落库
// =============================================================================

type cfcaPersisted struct {
	SignCertID string
	EncCertID  string
}

func (s *CFCAService) persist(
	ctx context.Context, req *CFCASignRequest,
	signCertPEM, encCertPEM, encKeyPEM, signKeyPEM []byte,
) (*cfcaPersisted, error) {
	bin := s.parser.OpensslBin()
	dirNo, err := s.layout.AllocServerDir()
	if err != nil {
		return nil, err
	}
	dir, _ := s.layout.ServerDir(dirNo)

	signPubSM3, err := PubkeySM3FromCertPEM(bin, string(signCertPEM))
	if err != nil {
		return nil, exception.New(exception.CodeInternalError, "签名证书公钥 SM3 失败", 500, nil)
	}
	encPubSM3, err := PubkeySM3FromCertPEM(bin, string(encCertPEM))
	if err != nil {
		return nil, exception.New(exception.CodeInternalError, "加密证书公钥 SM3 失败", 500, nil)
	}

	signCertAbs := filepath.Join(dir, signPubSM3+".cert.pem")
	encCertAbs := filepath.Join(dir, encPubSM3+".cert.pem")
	if err := os.WriteFile(signCertAbs, signCertPEM, 0640); err != nil {
		return nil, exception.New(exception.CodeInternalError, "写签名证书失败", 500, nil)
	}
	if err := os.WriteFile(encCertAbs, encCertPEM, 0640); err != nil {
		return nil, exception.New(exception.CodeInternalError, "写加密证书失败", 500, nil)
	}

	var signKeyAbs string
	if len(signKeyPEM) > 0 {
		signKeyAbs = filepath.Join(dir, signPubSM3+".key.pem")
		plainTmp := filepath.Join(dir, ".cfca.sign.key.plain.tmp")
		if err := os.WriteFile(plainTmp, signKeyPEM, 0600); err != nil {
			return nil, exception.New(exception.CodeInternalError, "写临时签名私钥失败", 500, nil)
		}
		defer os.Remove(plainTmp)
		if err := s.whitebox.EncryptFile(ctx, plainTmp, signKeyAbs); err != nil {
			return nil, exception.New(exception.CodeInternalError, "白盒加密签名私钥失败", 500, nil)
		}
		_ = os.Chmod(signKeyAbs, 0600)
	}

	encKeyAbs := filepath.Join(dir, encPubSM3+".key.pem")
	plainTmp := filepath.Join(dir, ".cfca.enc.key.plain.tmp")
	if err := os.WriteFile(plainTmp, encKeyPEM, 0600); err != nil {
		return nil, exception.New(exception.CodeInternalError, "写临时加密私钥失败", 500, nil)
	}
	defer os.Remove(plainTmp)
	if err := s.whitebox.EncryptFile(ctx, plainTmp, encKeyAbs); err != nil {
		return nil, exception.New(exception.CodeInternalError, "白盒加密加密私钥失败", 500, nil)
	}
	_ = os.Chmod(encKeyAbs, 0600)

	now := time.Now().UTC()
	baseID := newShortID()

	signDetail, _ := s.parser.Parse(signCertAbs)
	signCert := &models.Certificate{
		CertID:    "cfca-sign-" + baseID,
		CertType:  "dual_sign",
		CAID:      req.CAID,
		Algorithm: "SM2",
		CertPath:  signCertAbs,
		Status:    "VALID",
		NotBefore: now,
		NotAfter:  now.AddDate(0, 0, req.ValidityDays),
		CreatedAt: now,
	}
	if signDetail != nil {
		signCert.Subject = signDetail.Subject
		signCert.SubjectCN = cnFromSubject(signDetail.Subject)
		signCert.Issuer = signDetail.Issuer
		signCert.IssuerCN = cnFromSubject(signDetail.Issuer)
		signCert.Serial = signDetail.Serial
		signCert.Fingerprint = signDetail.Fingerprint
		signCert.PublicKeyAlgorithm = signDetail.PublicKeyAlgorithm
		signCert.SignatureAlgorithm = signDetail.SignatureAlgorithm
	}
	if signKeyAbs != "" {
		signCert.KeyPath = &signKeyAbs
		signCert.HasPrivateKey = true
	}
	if err := s.db.Create(signCert).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "签名证书落库失败", 500, nil)
	}

	encDetail, _ := s.parser.Parse(encCertAbs)
	encCert := &models.Certificate{
		CertID:    "cfca-enc-" + baseID,
		CertType:  "dual_enc",
		CAID:      req.CAID,
		Algorithm: "SM2",
		CertPath:  encCertAbs,
		Status:    "VALID",
		NotBefore: now,
		NotAfter:  now.AddDate(0, 0, req.ValidityDays),
		CreatedAt: now,
	}
	if encDetail != nil {
		encCert.Subject = encDetail.Subject
		encCert.SubjectCN = cnFromSubject(encDetail.Subject)
		encCert.Issuer = encDetail.Issuer
		encCert.IssuerCN = cnFromSubject(encDetail.Issuer)
		encCert.Serial = encDetail.Serial
		encCert.Fingerprint = encDetail.Fingerprint
		encCert.PublicKeyAlgorithm = encDetail.PublicKeyAlgorithm
		encCert.SignatureAlgorithm = encDetail.SignatureAlgorithm
	}
	encCert.KeyPath = &encKeyAbs
	encCert.HasPrivateKey = true
	if err := s.db.Create(encCert).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "加密证书落库失败", 500, nil)
	}

	log.Info().
		Str("dir_no", dirNo).
		Str("sign_cert_id", signCert.CertID).
		Str("enc_cert_id", encCert.CertID).
		Bool("sign_has_key", signKeyAbs != "").
		Msg("CFCA 双证签发完成")

	return &cfcaPersisted{SignCertID: signCert.CertID, EncCertID: encCert.CertID}, nil
}

// =============================================================================
// 解信封
// =============================================================================

func (s *CFCAService) DecodeEnvelope(
	ctx context.Context, req *CFCAQueryEnvelopeRequest,
) (*CFCAQueryEnvelopeResponse, error) {
	envelopeText := strings.TrimSpace(req.EncryptedEnvelope)
	if envelopeText == "" {
		return nil, exception.New(exception.CodeParamInvalid, "信封内容为空", 400, nil)
	}

	if err := s.ensureTmpKeyPair(); err != nil {
		return nil, exception.New(exception.CodeInternalError,
			"确保 CFCA 内置临时密钥失败: "+err.Error(), 500, nil)
	}
	if st, err := os.Stat(s.tmpKeyPath); err != nil || st.IsDir() {
		return nil, exception.New(exception.CodeInternalError,
			"CFCA 内置临时私钥不存在", 500, nil)
	}

	if strings.TrimSpace(req.TmpKeyPEM) != "" {
		log.Warn().Msg("cfca: 已忽略客户端传入的 tmp_key_pem，统一使用内置固定临时私钥")
	}

	var envBytes []byte
	if isCFCATextFormat(envelopeText) {
		envBytes = []byte(envelopeText)
	} else {
		decoded, err := base64.StdEncoding.DecodeString(envelopeText)
		if err == nil && len(decoded) > 0 {
			envBytes = decoded
		} else {
			envBytes = []byte(envelopeText)
		}
	}

	tmpDir, err := os.MkdirTemp("", "cfca-unseal-")
	if err != nil {
		return nil, exception.New(exception.CodeInternalError, "创建临时目录失败", 500, nil)
	}
	defer os.RemoveAll(tmpDir)

	envPath := filepath.Join(tmpDir, "env.bin")
	outPath := filepath.Join(tmpDir, "enc.key")
	if err := os.WriteFile(envPath, envBytes, 0600); err != nil {
		return nil, exception.New(exception.CodeInternalError, "写信封文件失败", 500, nil)
	}

	cmd := exec.CommandContext(ctx, s.envelopeBin, "unseal",
		"--tmp-key", s.tmpKeyPath, "--in", envPath, "--out", outPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdin = nil
	if err := cmd.Run(); err != nil {
		return nil, exception.New(exception.CodeParamInvalid,
			fmt.Sprintf("解信封失败: %v, stderr=%s", err, strings.TrimSpace(stderr.String())), 400, nil)
	}
	plain, err := os.ReadFile(outPath)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError, "读解密私钥失败", 500, nil)
	}
	return &CFCAQueryEnvelopeResponse{DecryptedKeyPEM: string(plain)}, nil
}

// isCFCATextFormat CFCA 5 段格式：前 80 字符是十进制（0-9），后面含逗号
func isCFCATextFormat(s string) bool {
	if len(s) < 80 {
		return false
	}
	for i := 0; i < 80; i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return false
		}
	}
	return strings.Contains(s, ",")
}

// =============================================================================
// 辅助
// =============================================================================

func (s *CFCAService) deriveCSRKeyPath(csr models.CSR) string {
	if csr.CSRPath == "" {
		return ""
	}
	base := filepath.Base(csr.CSRPath)
	if !strings.HasSuffix(base, ".req.csr") {
		return ""
	}
	dir := filepath.Dir(csr.CSRPath)
	pubSM3 := strings.TrimSuffix(base, ".req.csr")
	return filepath.Join(dir, pubSM3+".key.pem")
}

func pemDER(pemBytes []byte) []byte {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil
	}
	return block.Bytes
}
