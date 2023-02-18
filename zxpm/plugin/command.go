package plugin

type CommandInterface interface {
	List() []CommandDescriptor
	Run(tag string, args []string, opts map[string]string) int
}

type CommandDescriptor struct {
	Tag               string
	Parents           []string
	Use               string
	Short             string
	MinPositionalArgs int
	MaxPositionalargs int
	Options           []OptionDescriptor
}

type OptionType int

const (
	OptionBool OptionType = iota + 1
	OptionaString
)

type OptionDescriptor struct {
	Name      string
	Shorthand string
	Usage     string
	Default   string
	ValueType OptionType
}
