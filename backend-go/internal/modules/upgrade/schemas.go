package upgrade

// UpgradeManifest 升级包清单（由 build_upgrade_script.sh 生成、发布方签名）。
//
// 与 create_package/upgrade/manifest.json 保持一致。
type UpgradeManifest struct {
	SchemaVersion    string `json:"schema_version"`
	Name             string `json:"name"`
	FromVersion      string `json:"from_version"`
	ToVersion        string `json:"to_version"`
	MinPlatformVer   string `json:"min_platform_version"`
	Channel          string `json:"channel"`
	Requires         string `json:"requires"`
	CreatedAt        string `json:"created_at"`
	BackendArtifact  string `json:"backend_artifact"`  // backend.tar.gz
	FrontendArtifact string `json:"frontend_artifact"` // dist.war
	BackendSHA256    string `json:"backend_sha256"`
	FrontendSHA256   string `json:"frontend_sha256"`
	SignatureFile    string `json:"signature_file"` // manifest.json.sig
}

// Status 升级状态（供前端「推送升级提醒」与状态展示）。
type Status struct {
	CurrentVersion    string             `json:"current_version"`
	InstallRoot       string             `json:"install_root"`
	Pending           []*UpgradeManifest `json:"pending,omitempty"` // 检测到但未应用的可用升级
	LastUpgradeAt     string             `json:"last_upgrade_at,omitempty"`
	LastUpgradeResult string             `json:"last_upgrade_result,omitempty"` // ok / rolled_back / failed
	UpdateAvailable   bool               `json:"update_available"`
}

// PackageOut 单个升级包摘要。
type PackageOut struct {
	ID              string `json:"id"`
	FromVersion     string `json:"from_version"`
	ToVersion       string `json:"to_version"`
	Channel         string `json:"channel"`
	Requires        string `json:"requires"`
	CreatedAt       string `json:"created_at"`
	BackendArtifact string `json:"backend_artifact"`
	BackendSHA256   string `json:"backend_sha256"`
	FrontendSHA256  string `json:"frontend_sha256"`
	Verified        bool   `json:"verified"` // manifest 格式与文件存在性已校验
}

// RunReq 执行升级请求。
type RunReq struct {
	PackageID string `json:"package_id" binding:"required"`
}

// RollbackReq 回退请求。
type RollbackReq struct {
	ID string `json:"id" binding:"required"`
}
