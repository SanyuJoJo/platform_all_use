package crypto

import "encoding/asn1"

// OIDCFCATmpPublicKey 是 CFCA 规范中临时公钥的 PKCS#10 attributes OID。
// 注意：它位于 CSR 的 attributes 段，不是 extensions。
var OIDCFCATmpPublicKey = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 63}

// CFCASignRequest 与 SignCertRequest 字段对齐。
type CFCASignRequest struct {
	// ---------- CA 来源 ----------
	CASource      string `json:"ca_source,omitempty"` // local（默认）| manual
	CAID          string `json:"ca_id,omitempty"`
	CACertPEM     string `json:"ca_cert_pem,omitempty"`
	CAKeyPEM      string `json:"ca_key_pem,omitempty"`
	CAKeyPassword string `json:"ca_key_password,omitempty"`

	// ---------- CSR 来源 ----------
	CSRSource      string `json:"csr_source,omitempty"` // generate | upload | existing
	CSRID          string `json:"csr_id,omitempty"`
	CSRPEM         string `json:"csr_pem,omitempty"`
	CSRKeyPEM      string `json:"csr_key_pem,omitempty"`
	CSRKeyPassword string `json:"csr_key_password,omitempty"`

	// TmpPubPEM 已废弃：平台统一使用内置固定临时公钥（见 defaultCFCATmpDir）。
	// 保留字段仅为兼容旧调用，服务端忽略。
	TmpPubPEM string `json:"tmp_pub_pem,omitempty"`

	// ---------- 证书参数 ----------
	Subject      map[string]string `json:"subject,omitempty"`
	SAN          []string          `json:"san,omitempty"`
	ValidityDays int               `json:"validity_days,omitempty"`
	Domain       string            `json:"domain,omitempty"`

	// ---------- 私钥回传 ----------
	ReturnSignKey     bool   `json:"return_sign_key,omitempty"`
	KeyExportPassword string `json:"key_export_password,omitempty"`

	CipherOrder string `json:"cipher_order,omitempty"`
}

// CFCASignResponse 一键签发响应。
type CFCASignResponse struct {
	SignCertPEM       string `json:"sign_cert_pem"`
	EncCertPEM        string `json:"enc_cert_pem"`
	EncryptedEnvelope string `json:"encrypted_envelope"` // base64 DER
	SignCertID        string `json:"sign_cert_id,omitempty"`
	EncCertID         string `json:"enc_cert_id,omitempty"`
	// TmpKeyPEM 已废弃：平台统一使用内置固定临时密钥，不再逐次返回。
	TmpKeyPEM  string `json:"tmp_key_pem,omitempty"`
	SignKeyPEM string `json:"sign_key_pem,omitempty"` // ReturnSignKey=true 时返回
}

// CFCAQueryEnvelopeRequest 解信封请求。
//
// TmpKeyPEM 已废弃：平台统一使用内置固定临时私钥（见 defaultCFCATmpDir）。
// 保留字段仅为兼容旧前端，服务端忽略。
type CFCAQueryEnvelopeRequest struct {
	EncryptedEnvelope string `json:"encrypted_envelope" binding:"required"`
	TmpKeyPEM         string `json:"tmp_key_pem,omitempty"`
}

// CFCAQueryEnvelopeResponse 解信封响应。
type CFCAQueryEnvelopeResponse struct {
	DecryptedKeyPEM string `json:"decrypted_key_pem"`
}

// CFCAGenerateResult 是 CFCA P10 生成结果（供 csr.create 的 CFCA 分支使用）。
//
// 与 CFCASignResponse 区分：CFCASignResponse 是签发双证返回（含证书），
// 本结构仅包含生成 CSR 时需要的 PEM 数据。
type CFCAGenerateResult struct {
	CSRPEM     []byte
	SignKeyPEM []byte
	Subject    map[string]string
	SAN        []string
}

// cfcaResolvedCSR 内部结构：CSR 来源解析结果。
//
// TmpPubDER / TmpKeyPEM 字段已废弃（保留仅为结构兼容）：
//   - SignDual 一律从内置固定路径读取临时公钥；
//   - generateBundle 不再生成新的临时密钥对，因此不会填充 TmpPubDER / TmpKeyPEM。
type cfcaResolvedCSR struct {
	CSRPEM     []byte
	TmpPubDER  []byte // 已废弃
	SignKeyPEM []byte // 签名私钥 PEM（有则落盘/回传）
	TmpKeyPEM  []byte // 已废弃
}
