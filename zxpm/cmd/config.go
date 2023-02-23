package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/zostay/dev-tools/zxpm/config"
	"github.com/zostay/dev-tools/zxpm/plugin"
)

type loadedGoalsSet map[string]*cobra.Command

func (l loadedGoalsSet) mark(plugin, goalName string, goal *cobra.Command) {
	l[strings.Join([]string{plugin, goalName}, ":")] = goal
}
func (l loadedGoalsSet) is(plugin, goal string) bool {
	_, exists := l[strings.Join([]string{plugin, goal}, ":")]
	return exists
}
func (l loadedGoalsSet) get(plugin, goal string) *cobra.Command {
	return l[strings.Join([]string{plugin, goal}, ":")]
}

func getTasks(
	cfg *config.Config,
	name string,
	iface plugin.Interface,
) ([]plugin.TaskDescription, error) {
	ctx := context.Background()
	pctx := plugin.NewContext(cfg.ToKV("", "", name))
	ctx = plugin.InitializeContext(ctx, pctx)

	tasks, err := iface.Implements(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tasks for plugin %q: %w", name, err)
	}

	return tasks, nil
}

func getGoal(
	cfg *config.Config,
	name string,
	iface plugin.Interface,
	goalName string,
) (plugin.GoalDescription, error) {
	ctx := context.Background()
	pctx := plugin.NewContext(cfg.ToKV("", "", name))
	ctx = plugin.InitializeContext(ctx, pctx)

	goal, err := iface.Goal(ctx, goalName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch goal name %q from plugin %q: %w", goalName, name, err)
	}

	return goal, nil
}

func configureTasks(
	cfg *config.Config,
	plugins plugin.Clients,
	cmd *cobra.Command,
) error {
	ifaces, err := plugin.DispenseAll(plugins)
	if err != nil {
		return err
	}

	loadedGoals := make(loadedGoalsSet, 10)
	for name, iface := range ifaces {
		tasks, err := getTasks(cfg, name, iface)
		if err != nil {
			return err
		}

		for _, task := range tasks {
			goalCmd, err := configureGoalCommand(loadedGoals, cfg, ifaces, task)
			if err != nil {
				return err
			}

			goal, taskNames := config.GoalAndTaskNames(task.Name())

			parentCmd := goalCmd
			taskPath := "/" + goal
			for _, taskName := range taskNames {
				taskPath += "/" + taskName
				taskCmd := getSubcommand(parentCmd, taskName)
				if taskCmd != nil {
					taskCmd.Short += " " + task.Short()
					parentCmd = taskCmd
					continue
				}

				taskCmd = &cobra.Command{
					Use:   taskName,
					Short: task.Short(),
					// TODO Implement RunTask and set it on the cobra.Command
					// RunE:  RunTask(taskPath),
				}

				taskCmd.Flags().StringP("target", "t", "default", "the target configuration to use")
				taskCmd.Flags().StringSliceP("define", "d", nil, "define a variable in a=b format")

				parentCmd = taskCmd
			}
		}
	}

	return nil
}

func configureGoalCommand(
	loadedGoals loadedGoalsSet,
	cfg *config.Config,
	ifaces map[string]plugin.Interface,
	task plugin.TaskDescription,
) (*cobra.Command, error) {
	goalName, _ := config.GoalAndTaskNames(task.Name())
	if loadedGoals.is(task.Plugin(), goalName) {
		return loadedGoals.get(task.Plugin(), goalName), nil
	}

	pcfg := cfg.GetPluginByCommand(task.Plugin())
	if pcfg == nil {
		return nil, fmt.Errorf("plugin %q defines task %q which required plugin %q, which is not loaded", task.Plugin(), task.Name(), task.Plugin())
	}

	goalDesc, err := getGoal(cfg, pcfg.Name, ifaces[pcfg.Name], goalName)
	if err != nil {
		return nil, err
	}

	goalCmd := &cobra.Command{
		Use:     goalDesc.Name(),
		Short:   goalDesc.Short(),
		Aliases: goalDesc.Aliases(),
		// TODO Implement RunGoal and set it on the cobra.Command
		// RunE:    RunGoal(goalDesc),
	}

	goalCmd.Flags().StringP("target", "t", "default", "the target configuration to use")
	goalCmd.Flags().StringSliceP("define", "d", nil, "define a variable in a=b format")

	loadedGoals.mark(task.Plugin(), goalName, goalCmd)

	return goalCmd, nil
}

func getSubcommand(parentCmd *cobra.Command, name string) *cobra.Command {
	for _, cmd := range parentCmd.Commands() {
		if cmd.Name() == name {
			return cmd
		}
	}
	return nil
}
