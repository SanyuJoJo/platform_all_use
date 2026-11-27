package upgrade

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rs/zerolog/log"

	"backend-go/internal/config"
	"backend-go/internal/exception"
)

// Service 升级服务（编排层）。
//
// 设计：本服务只做「编排 + 校验」，不做任何文件系统级替换。
// 真正的槽位切换 / 回滚由 create_package/upgrade/apply.sh / rollback.sh
// 完成——它们是文件系统的单一事实来源，保证「动作尽量小、可回退」。
type Service struct {
	cfg *config.Config

	// updateDir 待升级包扫描目录（本地 / 内网预置）。
	// 该目录只允许读，绝不把路径拼到任意用户输入上。
	updateDir string
}

// NewService 创建升级服务。
//
// updateDir 必须是服务端编译期固定/配置校验过的目录，绝不由请求体决定。
func NewService(cfg *config.Config, updateDir string) *Service {
	if updateDir == "" {
		updateDir = filepath.Join(cfg.InstallRoot, "update")
	}
	return &Service{cfg: cfg, updateDir: updateDir}
}

// readState 读取安装根下的 state.json（由 install.sh 写入）。
func (s *Service) readState() (map[string]string, bool) {
	p := filepath.Join(s.cfg.InstallRoot, "state.json")
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, false
	}
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		log.Warn().Err(err).Str("path", p).Msg("upgrade: state.json 解析失败")
		return nil, false
	}
	return m, true
}

// currentVersion 返回当前安装版本（来自 state.json；读不到则回退到未知）。
func (s *Service) currentVersion() string {
	if st, ok := s.readState(); ok && st["version"] != "" {
		return st["version"]
	}
	return "unknown"
}

// listPackages 扫描 updateDir 下的 manifest.json，返回结构有效的包列表。
func (s *Service) listPackages() []*UpgradeManifest {
	if s.updateDir == "" {
		return nil
	}
	entries, err := os.ReadDir(s.updateDir)
	if err != nil {
		return nil
	}
	var out []*UpgradeManifest
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		mf := filepath.Join(s.updateDir, e.Name(), "manifest.json")
		data, err := os.ReadFile(mf)
		if err != nil {
			continue
		}
		var m UpgradeManifest
		if err := json.Unmarshal(data, &m); err != nil {
			continue
		}
		if m.ToVersion == "" || m.FromVersion == "" {
			continue
		}
		out = append(out, &m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ToVersion > out[j].ToVersion })
	return out
}

// Status 返回当前版本 + 可用升级清单（供「推送升级提醒」）。
func (s *Service) Status() (*Status, error) {
	st, hasState := s.readState()
	cur := s.currentVersion()

	status := &Status{
		CurrentVersion: cur,
		InstallRoot:    s.cfg.InstallRoot,
	}
	if hasState {
		status.LastUpgradeAt = st["last_upgrade_at"]
	}

	// 已成功应用/回滚的最新结果（若存在 state 记录）。
	if st != nil {
		if r := st["last_upgrade_result"]; r != "" {
			status.LastUpgradeResult = r
		}
	}

	// 扫描可用升级：to_version > current。
	var pending []*UpgradeManifest
	for _, m := range s.listPackages() {
		if !newerThan(m.ToVersion, cur) {
			continue
		}
		pending = append(pending, m)
	}
	if len(pending) > 0 {
		status.Pending = pending
		status.UpdateAvailable = true
	}
	return status, nil
}

// ListPackages 返回所有扫描到的包（含校验状态）。
func (s *Service) ListPackages() ([]*PackageOut, error) {
	var out []*PackageOut
	for _, m := range s.listPackages() {
		out = append(out, &PackageOut{
			ID:              filepath.Base(manifestDirFor(m)),
			FromVersion:     m.FromVersion,
			ToVersion:       m.ToVersion,
			Channel:         m.Channel,
			Requires:        m.Requires,
			CreatedAt:       m.CreatedAt,
			BackendArtifact: m.BackendArtifact,
			BackendSHA256:   m.BackendSHA256,
			FrontendSHA256:  m.FrontendSHA256,
			Verified:        m.MinPlatformVer != "" && m.Requires != "",
		})
	}
	return out, nil
}

// manifestDirFor 从 manifest 定位其所在目录名（用于列表展示）。
func manifestDirFor(m *UpgradeManifest) string {
	return m.Name + "-" + m.ToVersion
}

// Run 执行升级（动作尽量小：槽位切换 + restart.sh，绝无关机）。
//
// 只做最小前置校验，然后把实际执行交给 apply.sh。
func (s *Service) Run(pkgID string) (*Status, error) {
	if err := s.sanitizeID(pkgID); err != nil {
		return nil, err
	}
	script := filepath.Join(s.cfg.InstallRoot, "bin", "upgrade", "apply.sh")
	if !fileExists(script) {
		return nil, exception.New(exception.CodeInternalError,
			"升级脚本不存在：bin/upgrade/apply.sh", 500, nil)
	}
	out, err := runScript(script, "--pkg", pkgID, "--install-dir", s.cfg.InstallRoot)
	if err != nil {
		log.Error().Err(err).Str("pkg", pkgID).Msg("upgrade: apply.sh 执行失败")
		return nil, exception.New(exception.CodeInternalError,
			"升级执行失败: "+err.Error(), 500, nil)
	}
	log.Info().Str("pkg", pkgID).Msg("upgrade: 升级成功")
	return s.StatusWithResult(out, "ok")
}

// Rollback 回退到上一槽位。
func (s *Service) Rollback(id string) (*Status, error) {
	if err := s.sanitizeID(id); err != nil {
		return nil, err
	}
	script := filepath.Join(s.cfg.InstallRoot, "bin", "upgrade", "rollback.sh")
	if !fileExists(script) {
		return nil, exception.New(exception.CodeInternalError,
			"回退脚本不存在：bin/upgrade/rollback.sh", 500, nil)
	}
	out, err := runScript(script, "--id", id, "--install-dir", s.cfg.InstallRoot)
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("upgrade: rollback.sh 执行失败")
		return nil, exception.New(exception.CodeInternalError,
			"回退失败: "+err.Error(), 500, nil)
	}
	log.Info().Str("id", id).Msg("upgrade: 回退成功")
	return s.StatusWithResult(out, "rolled_back")
}

// StatusWithResult 追加最近一次操作结果到状态。
func (s *Service) StatusWithResult(output, result string) (*Status, error) {
	status, err := s.Status()
	if err != nil {
		return nil, err
	}
	status.LastUpgradeResult = result
	return status, nil
}

func (s *Service) sanitizeID(id string) error {
	if id == "" {
		return exception.New(exception.CodeParamInvalid, "id 不能为空", 400, nil)
	}
	if strings.ContainsAny(id, "/\\ ..") {
		return exception.New(exception.CodeParamInvalid, "非法的 id", 400, nil)
	}
	return nil
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

// runScript 执行外部升级脚本（受控路径 + 参数白名单）。
func runScript(script string, args ...string) (string, error) {
	cmd := exec.Command(script, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// newerThan 粗糙语义比较 x > y（x.y.z 三段）。不相等返回 false。
func newerThan(x, y string) bool {
	return compareVer(x, y) > 0
}

func compareVer(x, y string) int {
	px := parseVer(x)
	py := parseVer(y)
	for i := 0; i < 3; i++ {
		if px[i] != py[i] {
			if px[i] > py[i] {
				return 1
			}
			return -1
		}
	}
	return 0
}

func parseVer(v string) [3]int {
	var out [3]int
	parts := strings.SplitN(strings.TrimSpace(v), ".", 3)
	for i := 0; i < len(parts) && i < 3; i++ {
		n := 0
		for _, c := range parts[i] {
			if c < '0' || c > '9' {
				break
			}
			n = n*10 + int(c-'0')
		}
		out[i] = n
	}
	return out
}
