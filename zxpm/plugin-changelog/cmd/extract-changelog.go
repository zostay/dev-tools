package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/zostay/dev-tools/zxpm/changes"
	"github.com/zostay/dev-tools/zxpm/release"
)

var (
	ExtractChangelog = &cobra.Command{
		Use:   "extract <version>",
		Short: "extract the bullets for the changelog section for the given version",
		Args:  cobra.ExactArgs(1),
		Run:   ExtractChangelogMain,
	}
)

func ExtractChangelogMain(_ *cobra.Command, args []string) {
	r, err := changes.ExtractSection(release.GoEmailConfig.Changelog, args[0])
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to read changelog section: %v\n", err)
		os.Exit(1)
	}

	_, err = io.Copy(os.Stdout, r)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to read changelog data: %v\n", err)
		os.Exit(1)
	}
}