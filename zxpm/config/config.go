package config

import (
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"

	"github.com/zostay/dev-tools/zxpm/storage"
)

type Config struct {
	Properties storage.KV

	Goals   []GoalConfig
	Plugins []PluginConfig
}

type GoalConfig struct {
	Name string

	EnabledPlugins  []string
	DisabledPlugins []string

	Properties storage.KV

	Tasks   []TaskConfig
	Targets []TargetConfig
}

type PluginConfig struct {
	Name    string
	Command string

	Properties storage.KV
}

type TaskConfig struct {
	Name    string
	SubTask string

	EnabledPlugins  []string
	DisabledPlugins []string

	Properties storage.KV

	Targets []TargetConfig
	Tasks   []TaskConfig
}

type TargetConfig struct {
	Name string

	EnabledPlugins  []string
	DisabledPlugins []string

	Properties storage.KV
}

func Load(filename string, in io.Reader) (*Config, error) {
	var raw RawConfig
	fileBytes, err := io.ReadAll(in)
	if err != nil {
		return nil, err
	}

	file, diags := hclsyntax.ParseConfig(fileBytes, filename, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, diags
	}

	diags = gohcl.DecodeBody(file.Body, nil, &raw)
	if diags.HasErrors() {
		return nil, diags
	}

	return decodeRawConfig(&raw)
}

func GoalAndTaskNames(taskPath string) (string, []string) {
	taskPath = path.Clean(taskPath)
	if taskPath == "" || taskPath == "/" {
		return "", nil
	}

	taskParts := strings.Split(taskPath, "/")
	if taskParts[0] == "" {
		taskParts = taskParts[1:]
	}

	if len(taskParts) == 1 {
		return taskParts[0], nil
	}

	return taskParts[0], taskParts[1:]
}

func (c *Config) GetGoalFromPath(taskPath string) *GoalConfig {
	goalName, _ := GoalAndTaskNames(taskPath)
	return c.GetGoal(goalName)
}

func (c *Config) GetGoalAndTasks(taskPath string) (*GoalConfig, []*TaskConfig) {
	goalName, taskNames := GoalAndTaskNames(taskPath)

	goal := c.GetGoal(goalName)
	if goal == nil {
		panic(fmt.Sprintf("did not find goal %q", goalName))
	}

	tasks := make([]*TaskConfig, 0, len(taskNames))
	taskList := goal.Tasks

TaskLoop:
	for _, taskName := range taskNames {
		for j := range taskList {
			if taskList[j].Name == taskName {
				tasks = append(tasks, &taskList[j])
				taskList = taskList[j].Tasks
				continue TaskLoop
			}
		}
		break
	}
	return goal, tasks
}

func (c *Config) GetGoal(goalName string) *GoalConfig {
	for i := range c.Goals {
		if c.Goals[i].Name == goalName {
			return &c.Goals[i]
		}
	}
	return nil
}

func (c *Config) GetPlugin(pluginName string) *PluginConfig {
	for i := range c.Plugins {
		if c.Plugins[i].Name == pluginName {
			return &c.Plugins[i]
		}
	}
	return nil
}

func (c *Config) GetPluginByCommand(pluginCommand string) *PluginConfig {
	for i := range c.Plugins {
		if c.Plugins[i].Command == pluginCommand {
			return &c.Plugins[i]
		}
	}
	return nil
}

func (c *Config) ToKV(taskPath, targetName, pluginName string) storage.KV {
	var (
		goal   *GoalConfig
		tasks  []*TaskConfig
		plugin *PluginConfig
	)

	if taskPath != "" {
		goal, tasks = c.GetGoalAndTasks(taskPath)
	}
	if pluginName != "" {
		plugin = c.GetPlugin(pluginName)
	}

	layers := make([]storage.KV, 0, (len(tasks)+2)*2)
	layers = append(layers, c.Properties)
	if plugin != nil {
		layers = append(layers, plugin.Properties)
	}

	if goal != nil {
		layers = append(layers, goal.Properties)

		var target *TargetConfig
		if targetName != "" {
			target = goal.GetTarget(targetName)
		}

		if target != nil {
			layers = append(layers, target.Properties)
		}
	}

	for _, task := range tasks {
		layers = append(layers, task.Properties)

		var target *TargetConfig
		if targetName != "" {
			target = task.GetTarget(targetName)
		}

		if target != nil {
			layers = append(layers, target.Properties)
		}
	}

	if len(layers) == 0 {
		return storage.New()
	}

	return storage.Layers(layers...)
}

func (g *GoalConfig) GetTarget(targetName string) *TargetConfig {
	for i := range g.Targets {
		if g.Targets[i].Name == targetName {
			return &g.Targets[i]
		}
	}
	return nil
}

func (t *TaskConfig) GetTarget(targetName string) *TargetConfig {
	for i := range t.Targets {
		if t.Targets[i].Name == targetName {
			return &t.Targets[i]
		}
	}
	return nil
}
