package plugin_changelog

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/zostay/dev-tools/zxpm/changes"
	"github.com/zostay/dev-tools/zxpm/plugin"
)

type InfoChangelogTask struct {
	plugin.Boilerplate
}

func (t *InfoChangelogTask) ExtractChangelog(ctx context.Context) error {
	version := plugin.GetString(ctx, "info.version")
	r, err := changes.ExtractSection(Changelog(ctx), version)
	if err != nil {
		return fmt.Errorf("failed to read changelog section: %w", err)
	}

	_, err = io.Copy(os.Stdout, r)
	if err != nil {
		return fmt.Errorf("failed to read changelog data: %w", err)
	}

	return nil
}

func (t *InfoChangelogTask) Run(context.Context) (plugin.Operations, error) {
	return plugin.Operations{
		{
			Order:  50,
			Action: plugin.OperationFunc(t.ExtractChangelog),
		},
	}, nil
}
