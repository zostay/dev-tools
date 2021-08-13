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

type WebTarget struct {
	Type  WebTargetType
	Build []string
	Run   []string

	AddressMatch string `mapstructure:"address_match"`

	Watches  []FileWatch
	Dispatch []ProxyDispatch
}

type ProxyDispatch struct {
	Target string
	Path   string
}
