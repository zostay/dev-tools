package changelogImpl

import (
	"context"
	"fmt"
	"io"

	"github.com/zostay/dev-tools/zxpm/changes"
	"github.com/zostay/dev-tools/zxpm/plugin"
)

type InfoChangelogTask struct {
	plugin.TaskBoilerplate
}

func (t *InfoChangelogTask) ExtractChangelog(ctx context.Context) error {
	version := plugin.GetString(ctx, "info.version")
	r, err := changes.ExtractSection(Changelog(ctx), version)
	if err != nil {
		return fmt.Errorf("failed to read changelog section: %w", err)
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read changelog data: %w", err)
	}

	plugin.Set(ctx, "release.description", string(data))
	plugin.Set(ctx, "__export__.release.description", true)

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
