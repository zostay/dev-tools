package plugin

import (
	"context"
)

type pluginContextKey struct{}
type PluginContext struct {
	cleanup      []SimpleTask
	addFiles     []string
	globalConfig *Config
	properties   map[string]string
	changes      map[string]string
}

type SimpleTask func()

func NewPluginContext(
	globalConfig *Config,
) *PluginContext {
	return &PluginContext{
		cleanup:      make([]SimpleTask, 0, 10),
		addFiles:     make([]string, 0, 10),
		globalConfig: globalConfig,
		properties:   make(map[string]string, 10),
		changes:      make(map[string]string, 10),
	}
}

func (p *PluginContext) UpdateStorage(store map[string]string) {
	p.properties = make(map[string]string, len(store))
	for k, v := range store {
		p.properties[k] = v
	}
}

func (p *PluginContext) StorageChanges() map[string]string {
	return p.changes
}

func InitializeContext(ctx context.Context, pctx *PluginContext) context.Context {
	return context.WithValue(ctx, pluginContextKey{}, pctx)
}

func pluginContextFrom(ctx context.Context) *PluginContext {
	v := ctx.Value(pluginContextKey{})
	pctx, isPluginContext := v.(*PluginContext)
	if !isPluginContext {
		panic("context is missing plugin configuration")
	}
	return pctx
}

func ForCleanup(ctx context.Context, newCleaner SimpleTask) {
	pctx := pluginContextFrom(ctx)
	pctx.cleanup = append(pctx.cleanup, newCleaner)
}

func ListCleanupTasks(ctx context.Context) []SimpleTask {
	pctx := pluginContextFrom(ctx)
	tasks := make([]SimpleTask, len(pctx.cleanup))
	for i, f := range pctx.cleanup {
		tasks[len(tasks)-i-1] = f
	}
	return tasks
}

func ToAdd(ctx context.Context, newFile string) {
	pctx := pluginContextFrom(ctx)
	pctx.addFiles = append(pctx.addFiles, newFile)
}

func ListAdded(ctx context.Context) []string {
	pctx := pluginContextFrom(ctx)
	return pctx.addFiles
}

func GetConfig(ctx context.Context, key string) string {
	pctx := pluginContextFrom(ctx)
	return pctx.globalConfig.Get(key)
}

func Set(ctx context.Context, key, value string) {
	pctx := pluginContextFrom(ctx)
	pctx.changes[key] = value
}

func Get(ctx context.Context, key string) string {
	pctx := pluginContextFrom(ctx)
	if v, changeExists := pctx.changes[key]; changeExists {
		return v
	}
	return pctx.properties[key]
}
