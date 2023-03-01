package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/zostay/dev-tools/zxpm/plugin/master"
)

var runCmd = &cobra.Command{
	Use:   "run [ -t <target> ] *[ -d <key>=<value> ]",
	Short: "Execute the tasks to achieve the named goal",
}

func init() {
	runCmd.PersistentFlags().StringP("target", "t", "default", "the target configuration to use")
	runCmd.PersistentFlags().StringToStringP("define", "d", nil, "define a variable in a=b format")
	// TODO Figure out a contract that we can use when defining plugins to enforce dry-run that we can be relatively sure will always work...
	// runCmd.PersistentFlags().Bool("dry-run", false, "describe what would happen if the command run without doing it")
}

func RunGoal(
	e *master.InterfaceExecutor,
	group *master.TaskGroup,
) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		target, _ := cmd.Flags().GetString("target")
		e.SetTargetName(target)

		values, _ := cmd.Flags().GetStringToString("define")
		e.Define(values)

		err := e.ExecuteAllStages(ctx, group)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to execute tasks (%s): %v\n", group.TaskNames(), err)
			os.Exit(1)
		}

		return nil
	}
}
