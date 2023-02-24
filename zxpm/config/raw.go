package config

import (
	"fmt"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"

	"github.com/zostay/dev-tools/zxpm/storage"
)

type RawConfig struct {
	Properties cty.Value `hcl:"properties,optional"`

	Goals   []RawGoalConfig   `hcl:"goal,block"`
	Plugins []RawPluginConfig `hcl:"plugin,block"`
}

type RawGoalConfig struct {
	Name string `hcl:"name,label"`

	EnabledPlugins  []string `hcl:"enabled,optional"`
	DisabledPlugins []string `hcl:"disabled,optional"`

	Properties cty.Value `hcl:"properties,optional"`

	Tasks   []RawTaskConfig   `hcl:"task,block"`
	Targets []RawTargetConfig `hcl:"target,block"`
}

type RawPluginConfig struct {
	Name    string `hcl:"name,label"`
	Command string `hcl:"command,label"`

	Properties cty.Value `hcl:"properties,optional"`
}

type RawTaskConfig struct {
	Name    string `hcl:"name,label""`
	SubTask string `hcl:"subtask,label"`

	EnabledPlugins  []string `hcl:"enabled,optional"`
	DisabledPlugins []string `hcl:"disabled,optional"`

	Properties cty.Value `hcl:"properties,optional"`

	Targets []RawTargetConfig `hcl:"target,block"`
	Tasks   []RawTaskConfig   `hcl:"task,block"`
}

type RawTargetConfig struct {
	Name string `hcl:"name,label"`

	EnabledPlugins  []string `hcl:"enabled,optional"`
	DisabledPlugins []string `hcl:"disabled,optional"`

	Properties cty.Value `hcl:"properties,optional"`
}

func p(prefix, key string) string {
	return prefix + key + "."
}

func decodeRawProperties(prefix string, in cty.Value) (storage.KV, error) {
	if in.IsNull() {
		return storage.New(), nil
	}

	if !in.Type().IsObjectType() {
		return nil, fmt.Errorf("%s properties must be set to an key/value map", p(prefix, "properties"))
	}

	inm := in.AsValueMap()
	out := storage.New()

	for k, v := range inm {
		if v.Type().IsMapType() {
			dv, err := decodeRawProperties(p(prefix, k), v)
			if err != nil {
				return nil, fmt.Errorf("%s %w", p(prefix, k), err)
			}

			out.Set(k, dv)
			continue
		}

		if v.Type().IsCapsuleType() {
			ev := v.EncapsulatedValue()
			out.Set(k, ev)
		}

		if v.Type() == cty.Bool {
			var val bool
			_ = gocty.FromCtyValue(v, &val)
			out.Set(k, val)
		} else if v.Type() == cty.Number {
			var val float64
			_ = gocty.FromCtyValue(v, &val)
			out.Set(k, val)
		} else if v.Type() == cty.String {
			var val string
			_ = gocty.FromCtyValue(v, &val)
			out.Set(k, val)
		} else {
			return nil, fmt.Errorf("unknown value type for key %q", k)
		}
	}

	return out, nil
}

func decodeRawConfig(rc *RawConfig) (*Config, error) {
	props, err := decodeRawProperties("", rc.Properties)
	if err != nil {
		return nil, err
	}

	goals, err := decodeRawList[RawGoalConfig, GoalConfig]("", rc.Goals, decodeRawGoal)
	if err != nil {
		return nil, err
	}

	plugins, err := decodeRawList[RawPluginConfig, PluginConfig]("", rc.Plugins, decodeRawPlugin)
	if err != nil {
		return nil, err
	}

	return &Config{
		Properties: props,
		Goals:      goals,
		Plugins:    plugins,
	}, nil
}

func decodeRawList[In any, Out any](
	prefix string,
	rs []In,
	decoder func(string, *In) (*Out, error),
) ([]Out, error) {
	out := make([]Out, len(rs))
	for i := range rs {
		r := &rs[i]
		c, err := decoder(prefix, r)
		if err != nil {
			return nil, err
		}

		out[i] = *c
	}
	return out, nil
}

func decodeRawGoal(prefix string, in *RawGoalConfig) (*GoalConfig, error) {
	pn := p(prefix, in.Name)
	props, err := decodeRawProperties(pn, in.Properties)
	if err != nil {
		return nil, err
	}

	tasks, err := decodeRawList[RawTaskConfig, TaskConfig](pn, in.Tasks, decodeRawTask)
	if err != nil {
		return nil, err
	}

	targets, err := decodeRawList[RawTargetConfig, TargetConfig](pn, in.Targets, decodeRawTarget)
	if err != nil {
		return nil, err
	}

	return &GoalConfig{
		Name:            in.Name,
		EnabledPlugins:  in.EnabledPlugins,
		DisabledPlugins: in.DisabledPlugins,
		Properties:      props,
		Tasks:           tasks,
		Targets:         targets,
	}, nil
}

func decodeRawPlugin(prefix string, in *RawPluginConfig) (*PluginConfig, error) {
	pn := p(prefix, in.Name)
	props, err := decodeRawProperties(pn, in.Properties)
	if err != nil {
		return nil, err
	}

	return &PluginConfig{
		Name:       in.Name,
		Command:    in.Command,
		Properties: props,
	}, nil
}

func decodeRawTask(prefix string, in *RawTaskConfig) (*TaskConfig, error) {
	pn := p(prefix, in.Name)
	props, err := decodeRawProperties(pn, in.Properties)
	if err != nil {
		return nil, err
	}

	targets, err := decodeRawList[RawTargetConfig, TargetConfig](pn, in.Targets, decodeRawTarget)
	if err != nil {
		return nil, err
	}

	tasks, err := decodeRawList[RawTaskConfig, TaskConfig](pn, in.Tasks, decodeRawTask)
	if err != nil {
		return nil, err
	}

	return &TaskConfig{
		Name:            in.Name,
		SubTask:         in.SubTask,
		EnabledPlugins:  in.EnabledPlugins,
		DisabledPlugins: in.DisabledPlugins,
		Properties:      props,
		Targets:         targets,
		Tasks:           tasks,
	}, nil
}

func decodeRawTarget(prefix string, in *RawTargetConfig) (*TargetConfig, error) {
	pn := p(prefix, in.Name)
	props, err := decodeRawProperties(pn, in.Properties)
	if err != nil {
		return nil, err
	}

	return &TargetConfig{
		Name:            in.Name,
		EnabledPlugins:  in.EnabledPlugins,
		DisabledPlugins: in.DisabledPlugins,
		Properties:      props,
	}, nil
}
