package handler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yourorg/core-go/internal/audit"
	"github.com/yourorg/core-go/internal/config"
	"github.com/yourorg/core-go/internal/envelope"
	"github.com/yourorg/core-go/internal/openssl"
	"github.com/yourorg/core-go/internal/pathguard"
)

// DualCertCreateHandler dual_cert.create。
//
// ★ 正确流程：
//   1. 客户本地生成 CSR，私钥 s_pri 永不上传；
//   2. core 用 CSR 的公钥 s_pub 签发【签名证书】（不重新生成签名密钥对）；
//   3. core 内部生成加密密钥对 e_pri / e_pub，用 e_pub 签发【加密证书】；
//   4. 返回 sign_cert_path / enc_cert_path / enc_key_ref（一次性）。
//
// ★ 使用者 DN：
//   - 若 params.subject 提供，则签名证书和加密证书都使用该 DN
//     （签名证书通过 -subj 覆盖 CSR 中的默认 DN）；
//   - 若未提供，则签名证书沿用 CSR 的 DN，加密证书从 CSR 中提取相同的 DN，
//     保证两证书 DN 一致。
//
// ★ 输入 params：
//   csr_path        : 客户上传的 CSR 文件路径（core root 相对路径）—— 必填
//   ca_id           : CA 标识
//   ca_cert_path    : CA 证书路径
//   ca_key_ref      : CA 私钥 key_ref
//   subject         : 使用者 DN（可选；不填则用 CSR 里的）
//   san             : SAN 列表
//   validity_days   : 有效期
//
// ★ 输出 data：
//   sign_cert_path  : 签名证书路径
//   sign_cert_id    : 签名证书内部 ID
//   enc_cert_path   : 加密证书路径
//   enc_cert_id     : 加密证书内部 ID
//   enc_key_ref     : 加密私钥的一次性 key_ref（后端构建信封后立即删除）
//   chain_path      : 链文件路径
//   ★ 不返回 sign_key_ref —— 签名私钥 s_pri 只在客户端
type DualCertCreateHandler struct{}

func (h *DualCertCreateHandler) Execute(
	ctx context.Context, req *envelope.Request, cfg *config.Config,
) *envelope.Response {
	start := time.Now()

	signAlg := strParamDefault(req.Params, "sign_algorithm", "SM2")
	encAlg := strParamDefault(req.Params, "enc_algorithm", "SM2")
	tlcp := strParamDefault(req.Params, "tlcp_profile", "GB/T 38636-2020")

	if signAlg != "SM2" || encAlg != "SM2" || tlcp != "GB/T 38636-2020" {
		return envelope.NewErrorWithDetail("ALGORITHM_NOT_ALLOWED",
			"dual cert requires SM2+SM2+GB/T 38636-2020",
			map[string]interface{}{
				"sign_algorithm": signAlg,
				"enc_algorithm":  encAlg,
				"tlcp_profile":   tlcp,
			})
	}

	// ---- 必须提供 CSR ----
	csrPath := strParam(req.Params, "csr_path")
	if csrPath == "" {
		return genErr("INVALID_PARAM", "missing csr_path (客户 CSR 必须提供)")
	}

	caCertRel := strParam(req.Params, "ca_cert_path")
	caKeyRef := strParam(req.Params, "ca_key_ref")
	if caCertRel == "" || caKeyRef == "" {
		return genErr("INVALID_PARAM", "missing ca_cert_path or ca_key_ref")
	}

	days := intParam(req.Params, "validity_days", 365)
	if days <= 0 {
		days = 365
	}

	// ---- 路径解析 ----
	csrAbs, err := pathguard.Resolve(csrPath)
	if err != nil {
		return genErr("PATH_NOT_ALLOWED", "invalid csr_path: "+err.Error())
	}
	if _, err := os.Stat(csrAbs); err != nil {
		return genErr("INVALID_PARAM", "csr file not found: "+csrPath)
	}

	caCertAbs, err := pathguard.Resolve(caCertRel)
	if err != nil {
		return genErr("PATH_NOT_ALLOWED", err.Error())
	}
	if _, err := os.Stat(caCertAbs); err != nil {
		return genErr("CERT_NOT_FOUND", "ca cert not found")
	}

	tmpDir, err := mkTmpDir("dual-cert")
	if err != nil {
		return genErr("CORE_EXEC_FAILED", err.Error())
	}
	defer cleanupTmpDir(tmpDir)

	caKeyAbs, err := loadDecryptedKey(caKeyRef, tmpDir)
	if err != nil {
		return genErr("KEY_NOT_FOUND", err.Error())
	}

	client, err := openssl.NewClient()
	if err != nil {
		return genErr("CORE_EXEC_FAILED", err.Error())
	}

	digest := digestForCertAlg(ctx, client, caCertAbs)

	// ========================================================================
	// 步骤 0：确定使用者 DN
	//
	//   优先 params.subject（用户显式输入）；
	//   否则从 CSR 中提取。
	//
	//   userProvidedSubject=true 时：
	//     - 签名证书：通过 openssl x509 -subj 覆盖 CSR 里的 DN
	//     - 加密证书：用该 DN 生成加密 CSR
	//   userProvidedSubject=false 时：
	//     - 签名证书：直接用 CSR，不加 -subj
	//     - 加密证书：用从 CSR 提取的 DN
	// ========================================================================
	userProvidedSubject := false
	subjectArg := ""
	if subject, ok := subjectFromParams(req.Params); ok && len(subject) > 0 {
		subjectArg = subjectToOpenSSL(subject)
		userProvidedSubject = true
	}
	if subjectArg == "" {
		subjOut, err := client.Run(ctx, "req", "-in", csrAbs, "-noout", "-subject", "-nameopt", "RFC2253")
		if err == nil {
			subjStr := strings.TrimSpace(string(subjOut))
			subjStr = strings.TrimSpace(strings.TrimPrefix(subjStr, "subject="))
			if subjStr != "" {
				subjectArg = rfc2253ToSubj(subjStr)
			}
		}
	}
	if subjectArg == "" {
		return genErr("INVALID_PARAM", "cannot determine subject (既未提供 subject，也无法从 CSR 提取)")
	}

	// ========================================================================
	// 步骤 1：用 CSR 签发签名证书
	//
	// ★ 关键：openssl x509 -req 会从 CSR 中读取公钥并写入证书，
	//   因此签名证书公钥 = CSR 公钥 = 客户 s_pub。
	//   若用户提供了 DN，用 -subj 覆盖 CSR 里的 DN。
	// ========================================================================
	signCertID := newID("dual-sign")
	signCertRel := "data/certs/" + signCertID + ".pem"
	signCertAbs, err := pathguard.Resolve(signCertRel)
	if err != nil {
		return genErr("PATH_NOT_ALLOWED", err.Error())
	}
	if err := pathguard.EnsureDir(filepath.Dir(signCertAbs), 0750); err != nil {
		return genErr("CORE_EXEC_FAILED", err.Error())
	}

	signExt := filepath.Join(tmpDir, "sign.ext")
	signExtContent := "basicConstraints=CA:FALSE\n" +
		"keyUsage=critical,digitalSignature,nonRepudiation\n" +
		"extendedKeyUsage=clientAuth\n"
	if err := os.WriteFile(signExt, []byte(signExtContent), 0600); err != nil {
		return genErr("CORE_EXEC_FAILED", err.Error())
	}

	signArgs := []string{
		"x509", "-req",
		"-in", csrAbs,
		"-CA", caCertAbs,
		"-CAkey", caKeyAbs,
		"-CAcreateserial",
		"-out", signCertAbs,
		"-days", fmt.Sprintf("%d", days),
		"-extfile", signExt,
	}
	// ★ 如果用户显式提供了 subject，用 -subj 覆盖 CSR 里的 DN
	if userProvidedSubject {
		signArgs = append(signArgs, "-subj", subjectArg)
	}
	if digest != "" {
		signArgs = append(signArgs, digest)
	}
	if _, err := client.Run(ctx, signArgs...); err != nil {
		audit.Log(req, "dual_cert.create", start, "FAILED", "CORE_EXEC_FAILED")
		return genErr("CORE_EXEC_FAILED", "sign cert: "+err.Error())
	}
	_ = os.Chmod(signCertAbs, 0640)

	if _, err := os.Stat(signCertAbs); err != nil {
		return genErr("CORE_EXEC_FAILED", "sign cert not generated")
	}

	// ========================================================================
	// 步骤 2：内部生成加密密钥对（e_pri / e_pub）
	// ========================================================================
	encKeyPath := filepath.Join(tmpDir, "enc.key")
	if err := genPrivateKey(ctx, "SM2", req.Params, encKeyPath); err != nil {
		return genErr("CORE_EXEC_FAILED", "gen enc key: "+err.Error())
	}

	// 加密证书使用与签名证书相同的 DN
	encCSR := filepath.Join(tmpDir, "enc.csr")
	if _, err := client.Run(ctx, "req", "-new", "-key", encKeyPath,
		"-out", encCSR, "-subj", subjectArg); err != nil {
		return genErr("CORE_EXEC_FAILED", "enc csr: "+err.Error())
	}

	// ========================================================================
	// 步骤 3：用 CA 签发加密证书
	// ========================================================================
	encCertID := newID("dual-enc")
	encCertRel := "data/certs/" + encCertID + ".pem"
	encCertAbs, err := pathguard.Resolve(encCertRel)
	if err != nil {
		return genErr("PATH_NOT_ALLOWED", err.Error())
	}
	if err := pathguard.EnsureDir(filepath.Dir(encCertAbs), 0750); err != nil {
		return genErr("CORE_EXEC_FAILED", err.Error())
	}

	encExt := filepath.Join(tmpDir, "enc.ext")
	encExtContent := "basicConstraints=CA:FALSE\n" +
		"keyUsage=critical,keyEncipherment,keyAgreement\n" +
		"extendedKeyUsage=clientAuth\n"
	if err := os.WriteFile(encExt, []byte(encExtContent), 0600); err != nil {
		return genErr("CORE_EXEC_FAILED", err.Error())
	}

	encArgs := []string{
		"x509", "-req",
		"-in", encCSR,
		"-CA", caCertAbs,
		"-CAkey", caKeyAbs,
		"-CAcreateserial",
		"-out", encCertAbs,
		"-days", fmt.Sprintf("%d", days),
		"-extfile", encExt,
	}
	if digest != "" {
		encArgs = append(encArgs, digest)
	}
	if _, err := client.Run(ctx, encArgs...); err != nil {
		audit.Log(req, "dual_cert.create", start, "FAILED", "CORE_EXEC_FAILED")
		return genErr("CORE_EXEC_FAILED", "enc cert: "+err.Error())
	}
	_ = os.Chmod(encCertAbs, 0640)

	if _, err := os.Stat(encCertAbs); err != nil {
		return genErr("CORE_EXEC_FAILED", "enc cert not generated")
	}

	// ========================================================================
	// 步骤 4：链文件
	// ========================================================================
	chainRel := "data/certs/" + signCertID + "-chain.pem"
	chainAbs, _ := pathguard.Resolve(chainRel)
	if chainAbs != "" {
		signData, _ := os.ReadFile(signCertAbs)
		caData, _ := os.ReadFile(caCertAbs)
		_ = os.WriteFile(chainAbs, append(signData, caData...), 0640)
	}

	// ========================================================================
	// 步骤 5：返回结果
	// ========================================================================
	encPlain, _ := os.ReadFile(encKeyPath)
	encKeyRef, err := saveEncryptedKey(encPlain)
	if err != nil {
		return genErr("CORE_EXEC_FAILED", "save enc key: "+err.Error())
	}

	audit.Log(req, "dual_cert.create", start, "SUCCESS", "")

	return envelope.NewSuccess(map[string]interface{}{
		"sign_cert_path": signCertRel,
		"sign_cert_id":   signCertID,
		"enc_cert_path":  encCertRel,
		"enc_cert_id":    encCertID,
		"enc_key_ref":    encKeyRef,
		"chain_path":     chainRel,
	})
}

// rfc2253ToSubj 把 RFC2253 subject 转换为 openssl -subj 风格。
//
// 输入： "CN=Test,O=Example,C=CN"
// 输出： "/C=CN/O=Example/CN=Test"
func rfc2253ToSubj(rfc string) string {
	parts := splitRFC2253(rfc)
	kv := make(map[string]string)
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		idx := strings.Index(p, "=")
		if idx <= 0 {
			continue
		}
		k := strings.ToUpper(strings.TrimSpace(p[:idx]))
		v := strings.TrimSpace(p[idx+1:])
		kv[k] = v
	}
	order := []string{"C", "ST", "L", "O", "OU", "CN", "E", "SERIALNUMBER"}
	var sb strings.Builder
	used := map[string]bool{}
	for _, k := range order {
		if v, ok := kv[k]; ok && v != "" {
			sb.WriteString("/")
			sb.WriteString(k)
			sb.WriteString("=")
			sb.WriteString(v)
			used[k] = true
		}
	}
	for k, v := range kv {
		if used[k] || v == "" {
			continue
		}
		sb.WriteString("/")
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(v)
	}
	return sb.String()
}

// splitRFC2253 按未转义的逗号拆分 RFC2253 字符串。
func splitRFC2253(s string) []string {
	var (
		parts   []string
		current strings.Builder
		escaped bool
	)
	for _, r := range s {
		switch {
		case escaped:
			current.WriteRune(r)
			escaped = false
		case r == '\\':
			current.WriteRune(r)
			escaped = true
		case r == ',':
			parts = append(parts, current.String())
			current.Reset()
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}
