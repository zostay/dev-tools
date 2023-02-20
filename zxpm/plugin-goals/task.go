package plugin_goals

var _ pluginFullName.TaskDescription = &TaskDescription{}

type TaskDescription struct {
	plugin   string
	name     string
	short    string
	requires []string
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
