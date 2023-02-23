package goals

import (
	"github.com/zostay/dev-tools/zxpm/plugin"
)

var _ plugin.TaskDescription = &TaskDescription{}

type TaskDescription struct {
	plugin   string
	name     string
	short    string
	requires []string
}

func NewTaskDescription(plugin, name, short string, requires []string) *TaskDescription {
	return &TaskDescription{plugin, name, short, requires}
}

func (t *TaskDescription) Plugin() string {
	return t.plugin
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
