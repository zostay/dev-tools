package tools

import (
	"context"

	"github.com/zostay/dev-tools/pkg/config"
)

type pluginContextKey struct{}
type pluginContext struct {
	cleanup    []SimpleTask
	addFiles   []string
	config     *config.Config
	properties map[string]string
}

type SimpleTask func()

func newPluginContext(cfg *config.Config) *pluginContext {
	return &pluginContext{
		cleanup:    make([]SimpleTask, 0, 10),
		addFiles:   make([]string, 0, 10),
		config:     cfg,
		properties: make(map[string]string, 10),
	}
}

func InitializeContext(ctx context.Context, cfg *config.Config) context.Context {
	return context.WithValue(ctx, pluginContextKey{}, newPluginContext(cfg))
}

func pluginContextFrom(ctx context.Context) *pluginContext {
	v := ctx.Value(pluginContextKey{})
	pctx, isPluginContext := v.(*pluginContext)
	if !isPluginContext {
		panic("context is missing plugin configuration")
	}
	return pctx
}

func ForCleanup(ctx context.Context, newCleaner SimpleTask) {
	pctx := pluginContextFrom(ctx)
	pctx.cleanup = append(pctx.cleanup, newCleaner)
}

func ToAdd(ctx context.Context, newFile string) {
	pctx := pluginContextFrom(ctx)
	pctx.addFiles = append(pctx.addFiles, newFile)
}

func Set(ctx context.Context, key, value string) {
	pctx := pluginContextFrom(ctx)
	pctx.properties[key] = value
}
