package plugin_goals

import (
	"path"

	"github.com/zostay/dev-tools/zxpm/plugin"
)

var _ plugin.GoalDescription = &GoalDescription{}

type GoalDescription struct {
	name   string
	plugin string
	short  string
	alias  []string
}

func (g *GoalDescription) Task(name, short string, requires ...string) *TaskDescription {
	return &TaskDescription{
		plugin:   g.plugin,
		name:     g.TaskName(name),
		short:    short,
		requires: requires,
	}
}

func (g *GoalDescription) TaskName(name string) string {
	return path.Join(g.name, name)
}

func (g *GoalDescription) Name() string {
	return g.name
}

func (g *GoalDescription) Short() string {
	return g.short
}

func (g *GoalDescription) Aliases() []string {
	return g.alias
}
