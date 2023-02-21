package plugin_changelog

import (
	"context"
	"fmt"
	"os"

	"github.com/zostay/dev-tools/zxpm/changes"
	"github.com/zostay/dev-tools/zxpm/plugin"
)

const DefaultChangelog = "Changes.md"

func Changelog(ctx context.Context) string {
	cfg := plugin.ConfigFor(ctx, changelogPlugin)
	if cfg.IsSet("changelog") {
		return cfg.GetString("changelog")
	}
	return DefaultChangelog
}

func CheckPreRelease(ctx context.Context) bool {
	return plugin.GetBool(ctx, "lint.prerelease")
}

func CheckRelease(ctx context.Context) bool {
	return plugin.GetBool(ctx, "lint.release")
}

func CheckMode(ctx context.Context) changes.CheckMode {
	var mode changes.CheckMode
	switch {
	case CheckRelease(ctx):
		mode = changes.CheckRelease
	case CheckPreRelease(ctx):
		mode = changes.CheckPreRelease
	default:
		mode = changes.CheckStandard
	}
	return mode
}

// LintChangelog performs a check to ensure the changelog is ready for release.
func LintChangelog(ctx context.Context) error {
	changelog, err := os.Open(Changelog(ctx))
	if err != nil {
		return fmt.Errorf("unable to open Changes file: %w", err)
	}

	linter := changes.NewLinter(changelog, CheckMode(ctx))
	err = linter.Check()
	if err != nil {
		return err
	}

	return nil
}
