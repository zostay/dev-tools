package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"

	"github.com/zostay/dev-tools/zxpm/config"
	"github.com/zostay/dev-tools/zxpm/plugin"
	"github.com/zostay/dev-tools/zxpm/plugin/master"
	"github.com/zostay/dev-tools/zxpm/plugin/metal"
	"github.com/zostay/dev-tools/zxpm/storage"
)

type loadedGoalsSet map[string]*cobra.Command

func (l loadedGoalsSet) mark(goalName string, goal *cobra.Command) {
	l[goalName] = goal
}
func (l loadedGoalsSet) is(goal string) bool {
	_, exists := l[goal]
	return exists
}
func (l loadedGoalsSet) get(goal string) *cobra.Command {
	return l[goal]
}

func getTasks(
	cfg *config.Config,
	name string,
	iface plugin.Interface,
) ([]plugin.TaskDescription, error) {
	ctx := context.Background()
	rtStore := storage.New()
	pctx := plugin.NewConfigContext(rtStore, "", "", name, cfg)
	ctx = plugin.InitializeContext(ctx, pctx)

	tasks, err := iface.Implements(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tasks for plugin %q: %w", name, err)
	}

	return tasks, nil
}

func getGoal(
	m plugin.Interface,
	goalName string,
) (plugin.GoalDescription, error) {
	ctx := context.Background()

	goal, err := m.Goal(ctx, goalName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch goal name %q: %w", goalName, err)
	}

	return goal, nil
}

func configureTasks(
	cfg *config.Config,
	plugins metal.Clients,
	attachCmd *cobra.Command,
) error {
	ifaces, err := metal.DispenseAll(plugins)
	if err != nil {
		return err
	}

	m := master.NewInterface(cfg, ifaces)
	e := master.NewExecutor(m)

	ctx := context.Background()
	groups, err := e.TaskGroups(ctx)
	if err != nil {
		return err
	}

	cmds := make(map[string]*cobra.Command, len(groups))
	for _, group := range groups {
		cmd := configureGoalCommand(group, e)
		attachCmd.AddCommand(cmd)
		cmds[group.Tree] = cmd

		for _, sub := range group.SubTasks() {
			cmd := configureTaskCommand(sub, e)
			parent := path.Dir(sub.Tree)
			cmds[parent].AddCommand(cmd)
			cmds[sub.Tree] = cmd
		}
	}

	return nil
}

func configureTaskCommand(
	group *master.TaskGroup,
	e *master.InterfaceExecutor,
) *cobra.Command {
	shorts := make([]string, len(group.Tasks))
	for i, task := range group.Tasks {
		shorts[i] = task.Short()
	}

	return &cobra.Command{
		Use:   path.Base(group.Tree),
		Short: strings.Join(shorts, " "),
		RunE:  RunGoal(e, group),
	}
}

func configureGoalCommand(
	group *master.TaskGroup,
	e *master.InterfaceExecutor,
) *cobra.Command {
	return &cobra.Command{
		Use:     group.Goal.Name(),
		Short:   group.Goal.Short(),
		Aliases: group.Goal.Aliases(),
		RunE:    RunGoal(e, group),
	}
}
