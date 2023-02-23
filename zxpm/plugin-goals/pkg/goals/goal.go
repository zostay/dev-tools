package goals

import (
	"path"

	"github.com/zostay/dev-tools/zxpm/plugin"
)

var _ plugin.GoalDescription = &GoalDescription{}

type GoalDescription struct {
	name    string
	plugin  string
	short   string
	aliases []string
}

func NewGoalDescription(name, short string, aliases ...string) *GoalDescription {
	return &GoalDescription{name, "", short, aliases}
}

func NewGoalDescriptionForPlugin(name, plugin, short string, aliases ...string) *GoalDescription {
	return &GoalDescription{name, plugin, short, aliases}
}

func (g *GoalDescription) Task(name, short string, requires ...string) *TaskDescription {
	if g.plugin == "" {
		panic("cannot create task without plugin setting")
	}

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
	return g.aliases
}
