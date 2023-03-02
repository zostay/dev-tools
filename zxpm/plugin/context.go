package plugin

import (
	"context"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"

	"github.com/zostay/dev-tools/zxpm/config"
	"github.com/zostay/dev-tools/zxpm/storage"
)

type contextKey struct{}
type Context struct {
	logger     hclog.Logger
	cleanup    []SimpleTask
	addFiles   []string
	properties *storage.KVChanges
}

type SimpleTask func()

func NewContext(
	properties storage.KV,
) *Context {
	return &Context{
		logger:     hclog.L(),
		cleanup:    make([]SimpleTask, 0, 10),
		addFiles:   make([]string, 0, 10),
		properties: storage.WithChangeTracking(properties),
	}
}

func NewConfigContext(
	runtime storage.KV,
	taskName string,
	targetName string,
	pluginName string,
	cfg *config.Config,
) *Context {
	return NewContext(cfg.ToKV(runtime, taskName, targetName, pluginName))
}

func (p *Context) UpdateStorage(store map[string]any) {
	p.properties.Inner.Update(store)
}

func (p *Context) StorageChanges() map[string]string {
	changes := p.properties.ChangesStrings()
	p.properties.ClearChanges()
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

func Logger(ctx context.Context, withArgs ...interface{}) hclog.Logger {
	pctx := contextFrom(ctx)
	if len(withArgs) > 0 {
		return pctx.logger.With(withArgs...)
	}
	return pctx.logger
}

func ConfigName(o any) string {
	pkg := reflect.TypeOf(o).PkgPath()
	return strings.ReplaceAll(pkg, ".", "_")
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

func IsSet(ctx context.Context, key string) bool {
	pctx := contextFrom(ctx)
	return pctx.properties.IsSet(key)
}

func Set(ctx context.Context, key string, value any) {
	pctx := contextFrom(ctx)
	pctx.properties.Set(key, value)
}

func get[T any](ctx context.Context, key string, getter func(storage.KV, string) T) T {
	pctx := contextFrom(ctx)
	return getter(pctx.properties, key)
}

func Get(ctx context.Context, key string) any {
	return get(ctx, key, storage.KV.Get)
}

func GetBool(ctx context.Context, key string) bool {
	return get(ctx, key, storage.KV.GetBool)
}

func GetDuration(ctx context.Context, key string) time.Duration {
	return get(ctx, key, storage.KV.GetDuration)
}

func GetFloat64(ctx context.Context, key string) float64 {
	return get(ctx, key, storage.KV.GetFloat64)
}

func GetInt(ctx context.Context, key string) int {
	return get(ctx, key, storage.KV.GetInt)
}

func GetInt32(ctx context.Context, key string) int32 {
	return get(ctx, key, storage.KV.GetInt32)
}

func GetInt64(ctx context.Context, key string) int64 {
	return get(ctx, key, storage.KV.GetInt64)
}

func GetIntSlice(ctx context.Context, key string) []int {
	return get(ctx, key, storage.KV.GetIntSlice)
}

func GetString(ctx context.Context, key string) string {
	return get(ctx, key, storage.KV.GetString)
}

func GetStringMap(ctx context.Context, key string) map[string]any {
	return get(ctx, key, storage.KV.GetStringMap)
}

func GetStringMapString(ctx context.Context, key string) map[string]string {
	return get(ctx, key, storage.KV.GetStringMapString)
}

func GetStringMapStringSlice(ctx context.Context, key string) map[string][]string {
	return get(ctx, key, storage.KV.GetStringMapStringSlice)
}

func GetStringSlice(ctx context.Context, key string) []string {
	return get(ctx, key, storage.KV.GetStringSlice)
}

func GetTime(ctx context.Context, key string) time.Time {
	return get(ctx, key, storage.KV.GetTime)
}

func GetUint(ctx context.Context, key string) uint {
	return get(ctx, key, storage.KV.GetUint)
}

func GetUint16(ctx context.Context, key string) uint16 {
	return get(ctx, key, storage.KV.GetUint16)
}

func GetUint32(ctx context.Context, key string) uint32 {
	return get(ctx, key, storage.KV.GetUint32)
}

func GetUint64(ctx context.Context, key string) uint64 {
	return get(ctx, key, storage.KV.GetUint64)
}

func KV(ctx context.Context) storage.KV {
	pctx := contextFrom(ctx)
	return pctx.properties
}

func UpdateStrings(ctx context.Context, changes map[string]string) {
	pctx := contextFrom(ctx)
	pctx.properties.UpdateStrings(changes)
}
