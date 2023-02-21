package plugin_changelog

import (
	"context"
	"fmt"
	"io"

	"github.com/zostay/dev-tools/zxpm/changes"
	"github.com/zostay/dev-tools/zxpm/plugin"
	"github.com/zostay/dev-tools/zxpm/release"
)

type ReleaseFinishTask struct {
	plugin.Boilerplate

	Version   string
	Changelog string
}

// CaptureChangesInfo loads the bullets for the changelog section relevant to
// this release into the process configuration for use when creating the release
// later.
func (f *ReleaseFinishTask) CaptureChangesInfo(ctx context.Context) error {
	vstring := "v" + f.Version
	cr, err := changes.ExtractSection(f.Changelog, vstring)
	if err != nil {
		return fmt.Errorf("unable to get log of changes: %w", err)
	}

	chgs, err := io.ReadAll(cr)
	if err != nil {
		return fmt.Errorf("unable to read log of changes: %w", err)
	}

	plugin.Set(ctx, release.ValueDescription, string(chgs))

	return nil
}

func (f *ReleaseFinishTask) Check(ctx context.Context) error {
	if f.Changelog == "" {
		return fmt.Errorf("missing changelog location in paths config")
	}

	return f.CaptureChangesInfo(ctx)
}
