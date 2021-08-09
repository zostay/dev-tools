package config

type Web struct {
	Targets map[string]WebTarget
}

type WebTargetType string

const (
	APITarget     WebTargetType = "api"
	WebSiteTarget WebTargetType = "static"
	DockerTarget  WebTargetType = "docker"
)

type WebTarget struct {
	Type  WebTargetType
	Build []string
	Run   []string
	Port  int

	Watches []FileWatch
}
