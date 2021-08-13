package config

type Web struct {
	Targets map[string]WebTarget
}

type WebTargetType string

const (
	// ServerTarget commands are commands that need direct supervision to run
	// and require a separate build process whenever changes are detected. A
	// typical example is a Go API server.
	ServerTarget WebTargetType = "server"

	// StaticTarget commands are commands that combine the build and file
	// serving capabilities into a self-sufficient system.
	StaticTarget WebTargetType = "static"

	// DockerTarget commands are commands that run as docker processes and
	// docker commands are used to build and run them.
	DockerTarget WebTargetType = "docker"

	// FrontendTarget is an automatically configured front-end tool for acting
	// as a reverse proxy for development.
	FrontendTarget WebTargetType = "frontend"
)

type AddrFmt string

const (
	// AddrFmtHostPort is the format for host:port addresses. It needs to be
	// split and the host needs to be normalized (i.e., ::, is treated as ::1
	// and 0.0.0.0 is treated as 127.0.0.1, etc.).
	AddrFmtHostPort AddrFmt = "host:port"

	// AddrFmtURL is the format for full URL addresses.
	AddrFmtURL AddrFmt = "url"
)

type WebTarget struct {
	Type  WebTargetType
	Build []string
	Run   []string

	AddressMatch  string  `mapstructure:"address_match"`
	AddressFormat AddrFmt `mapstructure:"address_format"`

	OpenBrowser bool `mapstructure:"open_browser"`

	Watches  []FileWatch
	Dispatch []ProxyDispatch
}

type ProxyDispatch struct {
	Target string
	Path   string
}
