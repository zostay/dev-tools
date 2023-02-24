package goals

import (
	"github.com/zostay/dev-tools/zxpm/plugin"
)

var _ plugin.TaskDescription = &TaskDescription{}

type TaskDescription struct {
	name     string
	short    string
	requires []string
}

func NewTaskDescription(name, short string, requires []string) *TaskDescription {
	return &TaskDescription{name, short, requires}
}

func (t *TaskDescription) Name() string {
	return t.name
}

func (t *TaskDescription) Short() string {
	return t.short
}

func (t *TaskDescription) Requires() []string {
	return t.requires
}
