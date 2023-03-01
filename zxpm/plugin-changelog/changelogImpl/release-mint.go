package changelogImpl

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/zostay/dev-tools/zxpm/plugin"
)

type ReleaseMintTask struct {
	plugin.TaskBoilerplate
}

// FixupChangelog alters the changelog to prepare it for release.
func (s *ReleaseMintTask) FixupChangelog(ctx context.Context) error {
	r, err := os.Open(Changelog(ctx))
	if err != nil {
		return fmt.Errorf("unable to open %s: %w", Changelog(ctx), err)
	}

	newChangelog := Changelog(ctx) + ".new"

	w, err := os.Create(newChangelog)
	if err != nil {
		return fmt.Errorf("unable to create %s: %w", newChangelog, err)
	}

	plugin.ForCleanup(ctx, func() { _ = os.Remove(newChangelog) })

	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if line == "WIP" || line == "WIP  TBD" {
			version := plugin.GetString(ctx, "release.version")
			today := plugin.GetString(ctx, "release.date")
			_, _ = fmt.Fprintf(w, "v%s  %s\n", version, today)
		} else {
			_, _ = fmt.Fprintln(w, line)
		}
	}

	_ = r.Close()
	err = w.Close()
	if err != nil {
		return fmt.Errorf("unable to close %s: %w", newChangelog, err)
	}

	err = os.Rename(newChangelog, Changelog(ctx))
	if err != nil {
		return fmt.Errorf("unable to overwrite %s with %s: %w", Changelog(ctx), newChangelog, err)
	}

	plugin.ToAdd(ctx, Changelog(ctx))

	plugin.Logger(ctx,
		"changelog", Changelog(ctx),
	).Info("Applied changes to changelog to fixup for release.")

	return nil
}

func (s *ReleaseMintTask) Check(ctx context.Context) error {
	return LintChangelog(ctx)
}

func (s *ReleaseMintTask) Run(context.Context) (plugin.Operations, error) {
	return plugin.Operations{
		{
			Order:  50,
			Action: plugin.OperationFunc(s.FixupChangelog),
		},
		{
			Order:  55,
			Action: plugin.OperationFunc(LintChangelog),
		},
	}, nil
}
