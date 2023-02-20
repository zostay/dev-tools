package plugin_goals

var _ pluginFullName.GoalDescription = &GoalDescription{}

type GoalDescription struct {
	name   string
	plugin string
	short  string
	alias  []string
}

func (g *GoalDescription) Task(name, short string, requires ...string) *TaskDescription {
	return &TaskDescription{
		plugin:   g.plugin,
		name:     name,
		short:    short,
		requires: requires,
	}
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
