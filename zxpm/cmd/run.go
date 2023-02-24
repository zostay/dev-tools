package cmd

import "github.com/spf13/cobra"

var runCmd = &cobra.Command{
	Use:   "run [ -t <target> ] *[ -d <key>=<value> ]",
	Short: "Execute the tasks to achieve the named goal",
}

func init() {
	runCmd.Flags().StringP("target", "t", "default", "the target configuration to use")
	runCmd.Flags().StringSliceP("define", "d", nil, "define a variable in a=b format")
}
