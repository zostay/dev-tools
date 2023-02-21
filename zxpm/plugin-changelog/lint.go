package plugin_changelog

import (
	"context"

	"github.com/zostay/dev-tools/zxpm/plugin"
)

type LintChangelogTask struct {
	plugin.Boilerplate
}

func (t *LintChangelogTask) Run(ctx context.Context) (plugin.Operations, error) {
	return plugin.Operations{
		{
			Order:  50,
			Action: plugin.OperationFunc(LintChangelog),
		},
	}, nil
}
