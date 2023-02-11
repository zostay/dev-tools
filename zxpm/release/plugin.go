package release

import (
	"time"

	"github.com/coreos/go-semver/semver"
)

const StartTask = "start-release"
const FinishTask = "finish-release"

type TaskConfig struct {
	Version string
	Now     time.Time
}

func (t *TaskConfig) SemVer() *semver.Version {
	return semver.New(t.Version)
}
