package contract

// Manifest 描述插件的元信息，主应用通过扫描插件目录下的 manifest.json 识别插件。
type Manifest struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
	Author      string `json:"author,omitempty"`

	Mode    string `json:"mode"`    // "managed" | "standalone"
	Runtime string `json:"runtime"` // "go-plugin" | "http"

	// —— 路由映射：外部 URL → 本插件 ——
	Routes []Route `json:"routes,omitempty"`

	// —— managed 启动参数 ——
	Entrypoint string            `json:"entrypoint,omitempty"`
	ExecArgs   []string          `json:"exec_args,omitempty"`
	WorkDir    string            `json:"work_dir,omitempty"`
	Env        map[string]string `json:"env,omitempty"`

	// —— HTTP 通信参数 ——
	Host         string `json:"host,omitempty"`
	Port         int    `json:"port,omitempty"`
	HealthPath   string `json:"health_path,omitempty"`
	InitPath     string `json:"init_path,omitempty"`
	ExecutePath  string `json:"execute_path,omitempty"`
	ShutdownPath string `json:"shutdown_path,omitempty"`

	Checksum string `json:"checksum,omitempty"`
}

// Route 声明一条 URL 到插件的映射。
type Route struct {
	// Path 路径。以 "/" 结尾表示前缀匹配，否则精确匹配。
	// 例："/api/echo/"  命中 /api/echo/a、/api/echo/b
	//     "/api/echo"   只命中 /api/echo
	Path string `json:"path"`

	// Methods 允许的 HTTP 方法，空表示所有方法。
	Methods []string `json:"methods,omitempty"`

	// Priority 越大越优先，用于多个插件声明重叠路径时决定归属。
	Priority int `json:"priority,omitempty"`

	// StripPrefix 转发前是否剥离 Path 前缀。
	// 默认 false，表示插件收到的路径是原始 URL。
	// true 时，代理会把 /api/echo/foo 转发为 /foo。
	StripPrefix bool `json:"strip_prefix,omitempty"`
}

// MatchMethod 判断该路由是否允许给定 HTTP 方法。
func (r Route) MatchMethod(method string) bool {
	if len(r.Methods) == 0 {
		return true
	}
	for _, m := range r.Methods {
		if m == method {
			return true
		}
	}
	return false
}

// MatchPath 判断给定路径是否命中该路由。
func (r Route) MatchPath(path string) bool {
	if r.Path == "" {
		return false
	}
	if r.Path[len(r.Path)-1] == '/' {
		// 前缀匹配
		return len(path) >= len(r.Path) && path[:len(r.Path)] == r.Path
	}
	// 精确匹配
	return path == r.Path
}

// ApplyDefaults 填充 manifest 的默认值。
func (m *Manifest) ApplyDefaults() {
	if m.Host == "" {
		m.Host = "127.0.0.1"
	}
	if m.HealthPath == "" {
		m.HealthPath = "/health"
	}
	if m.InitPath == "" {
		m.InitPath = "/initialize"
	}
	if m.ExecutePath == "" {
		m.ExecutePath = "/execute"
	}
	if m.ShutdownPath == "" {
		m.ShutdownPath = "/shutdown"
	}
}

// HTTPBaseURL 返回 HTTP 插件的基础 URL。
func (m *Manifest) HTTPBaseURL() string {
	return "http://" + m.Host + ":" + itoa(m.Port)
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
