package changelog

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/zostay/dev-tools/zxpm/changes"
	"github.com/zostay/dev-tools/zxpm/plugin"
	"github.com/zostay/dev-tools/zxpm/plugin/tools"
)

type ReleaseStartTask struct {
	plugin.Boilerplate

	Version string
	Today   string

	Changelog string
}

// LintChangelog performs a check to ensure the changelog is ready for release.
func (s *ReleaseStartTask) LintChangelog(mode changes.CheckMode) error {
	changelog, err := os.Open(s.Changelog)
	if err != nil {
		return fmt.Errorf("unable to open Changes file: %w", err)
	}

	linter := changes.NewLinter(changelog, mode)
	err = linter.Check()
	if err != nil {
		return err
	}

	return nil
}

// FixupChangelog alters the changelog to prepare it for release.
func (s *ReleaseStartTask) FixupChangelog(ctx context.Context) error {
	r, err := os.Open(s.Changelog)
	if err != nil {
		return fmt.Errorf("unable to open %s: %w", s.Changelog, err)
	}

	newChangelog := s.Changelog + ".new"

	w, err := os.Create(newChangelog)
	if err != nil {
		return fmt.Errorf("unable to create %s: %w", newChangelog, err)
	}

	tools.ForCleanup(ctx, func() { _ = os.Remove(newChangelog) })

	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if line == "WIP" || line == "WIP  TBD" {
			_, _ = fmt.Fprintf(w, "v%s  %s\n", s.Version, s.Today)
		} else {
			_, _ = fmt.Fprintln(w, line)
		}
	}

	_ = r.Close()
	err = w.Close()
	if err != nil {
		return fmt.Errorf("unable to close %s: %w", newChangelog, err)
	}

	err = os.Rename(newChangelog, s.Changelog)
	if err != nil {
		return fmt.Errorf("unable to overwrite %s with %s: %w", s.Changelog, newChangelog, err)
	}

	tools.ToAdd(ctx, s.Changelog)

	return nil
}

func (s *ReleaseStartTask) Check(_ context.Context) error {
	if s.Changelog == "" {
		return fmt.Errorf("missing changelog location in paths config")
	}

	return s.LintChangelog(changes.CheckPreRelease)
}

func (s *ReleaseStartTask) Begin() plugin.Operations {
	return plugin.Operations{
		{
			Order:  50,
			Action: s.FixupChangelog,
		},
		{
			Order: 55,
			Action: func(context.Context) error {
				return s.LintChangelog(changes.CheckRelease)
			},
		},
	}
}
