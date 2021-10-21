package main

import (
	"github.com/spf13/cobra"
	"github.com/zostay/dev-tools/cmd/zxconfig/cmd"
)

// main runs the zxconfig command.
func main() {
	err := cmd.Execute()
	cobra.CheckErr(err)
}
