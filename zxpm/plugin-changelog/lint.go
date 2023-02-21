package plugin_changelog

import (
	"context"
	"fmt"
	"os"

	"github.com/zostay/dev-tools/zxpm/changes"
	"github.com/zostay/dev-tools/zxpm/plugin"
)

type LintChangelogTask struct {
	plugin.Boilerplate

	Changelog string
}

func (t *LintChangelogTask) LintChangelog(ctx context.Context) error {
	checkPreRelease := plugin.GetBool(ctx, "lint.prerelease")
	checkRelease := plugin.GetBool(ctx, "lint.release")

	changelog, err := os.Open(t.Changelog)
	if err != nil {
		return fmt.Errorf("unable to open %q file: %w", t.Changelog, err)
	}

	var mode changes.CheckMode
	switch {
	case checkRelease:
		mode = changes.CheckRelease
	case checkPreRelease:
		mode = changes.CheckPreRelease
	default:
		mode = changes.CheckStandard
	}

	linter := changes.NewLinter(changelog, mode)
	err = linter.Check()
	if err != nil {
		return err
	}

	return nil
}

func (t *LintChangelogTask) Run(ctx context.Context) (plugin.Operations, error) {
	return plugin.Operations{
		{
			Order:  50,
			Action: plugin.OperationFunc(t.LintChangelog),
		},
	}, nil
}
