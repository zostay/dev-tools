package cmd

import (
	"github.com/spf13/cobra"

	"github.com/zostay/dev-tools/zxpm/plugin/master"
)

var runCmd = &cobra.Command{
	Use:   "run [ -t <target> ] *[ -d <key>=<value> ]",
	Short: "Execute the tasks to achieve the named goal",
}

func init() {
	runCmd.Flags().StringP("target", "t", "default", "the target configuration to use")
	runCmd.Flags().StringToStringP("define", "d", nil, "define a variable in a=b format")
}

func RunGoal(
	e *master.InterfaceExecutor,
	group *master.TaskGroup,
) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// ctx := context.Background()

		target, _ := cmd.Flags().GetString("target")
		e.SetTargetName(target)

		values, _ := cmd.Flags().GetStringToString("define")
		e.Define(values)

		// TODO Execute task phases here...
		// err := e.Execute(ctx, group)
		//
		// if err != nil {
		// 	_, _ = fmt.Fprintf(os.Stderr, "failed to execute tasks: %v\n", err)
		// 	os.Exit(1)
		// }

		return nil
	}
}
