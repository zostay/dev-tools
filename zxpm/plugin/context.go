package plugin

import (
	"context"

	"github.com/zostay/dev-tools/pkg/config"
)

type contextKey struct{}
type Context struct {
	cleanup      []SimpleTask
	addFiles     []string
	globalConfig *config.Config
	properties   map[string]string
	changes      map[string]string
}

type SimpleTask func()

func NewPluginContext(
	globalConfig *config.Config,
) *Context {
	return &Context{
		cleanup:      make([]SimpleTask, 0, 10),
		addFiles:     make([]string, 0, 10),
		globalConfig: globalConfig,
		properties:   make(map[string]string, 10),
		changes:      make(map[string]string, 10),
	}
}

func (p *Context) UpdateStorage(store map[string]string) {
	p.properties = make(map[string]string, len(store))
	for k, v := range store {
		p.properties[k] = v
	}
}

func (p *Context) StorageChanges() map[string]string {
	var changes map[string]string
	changes, p.changes = p.changes, make(map[string]string, 10)
	return changes
}

func InitializeContext(ctx context.Context, pctx *Context) context.Context {
	return context.WithValue(ctx, contextKey{}, pctx)
}

func contextFrom(ctx context.Context) *Context {
	v := ctx.Value(contextKey{})
	pctx, isPluginContext := v.(*Context)
	if !isPluginContext {
		panic("context is missing plugin configuration")
	}
	return pctx
}

func ForCleanup(ctx context.Context, newCleaner SimpleTask) {
	pctx := contextFrom(ctx)
	pctx.cleanup = append(pctx.cleanup, newCleaner)
}

func ListCleanupTasks(ctx context.Context) []SimpleTask {
	pctx := contextFrom(ctx)
	tasks := make([]SimpleTask, len(pctx.cleanup))
	for i, f := range pctx.cleanup {
		tasks[len(tasks)-i-1] = f
	}
	return tasks
}

func ToAdd(ctx context.Context, newFile string) {
	pctx := contextFrom(ctx)
	pctx.addFiles = append(pctx.addFiles, newFile)
}

func ListAdded(ctx context.Context) []string {
	pctx := contextFrom(ctx)
	return pctx.addFiles
}

func GetConfig(ctx context.Context, key string) string {
	pctx := contextFrom(ctx)
	return pctx.globalConfig.Get(key)
}

func Set(ctx context.Context, key, value string) {
	pctx := contextFrom(ctx)
	pctx.changes[key] = value
}

func Get(ctx context.Context, key string) string {
	pctx := contextFrom(ctx)
	if v, changeExists := pctx.changes[key]; changeExists {
		return v
	}
	return pctx.properties[key]
}
