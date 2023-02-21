package plugin

import (
	"context"
	"reflect"
	"strings"
	"time"

	"github.com/zostay/dev-tools/zxpm/storage"
)

type contextKey struct{}
type Context struct {
	cleanup      []SimpleTask
	addFiles     []string
	globalConfig storage.KV
	properties   storage.KV
	changes      storage.KV
}

type SimpleTask func()

func NewContext(
	globalConfig storage.KV,
) *Context {
	return &Context{
		cleanup:      make([]SimpleTask, 0, 10),
		addFiles:     make([]string, 0, 10),
		globalConfig: globalConfig,
		properties:   storage.New(),
		changes:      storage.New(),
	}
}

func (p *Context) UpdateStorage(store map[string]any) {
	p.properties.Update(store)
}

func (p *Context) StorageChanges() map[string]any {
	changes := p.changes.AllSettings()
	p.changes.Clear()
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

func getc[T any](ctx context.Context, key string, getter func(storage.KV, string) T) T {
	pctx := contextFrom(ctx)
	return getter(pctx.globalConfig, key)
}

func ConfigFor(ctx context.Context, key string) storage.KV {
	return getc(ctx, key, storage.KV.Sub)
}

func GetConfig(ctx context.Context, key string) any {
	return getc(ctx, key, storage.KV.Get)
}

func GetConfigBool(ctx context.Context, key string) bool {
	return getc(ctx, key, storage.KV.GetBool)
}

func GetConfigDuration(ctx context.Context, key string) time.Duration {
	return getc(ctx, key, storage.KV.GetDuration)
}

func GetConfigFloat64(ctx context.Context, key string) float64 {
	return getc(ctx, key, storage.KV.GetFloat64)
}

func GetConfigInt(ctx context.Context, key string) int {
	return getc(ctx, key, storage.KV.GetInt)
}

func GetConfigInt32(ctx context.Context, key string) int32 {
	return getc(ctx, key, storage.KV.GetInt32)
}

func GetConfigInt64(ctx context.Context, key string) int64 {
	return getc(ctx, key, storage.KV.GetInt64)
}

func GetConfigIntSlice(ctx context.Context, key string) []int {
	return getc(ctx, key, storage.KV.GetIntSlice)
}

func GetConfigString(ctx context.Context, key string) string {
	return getc(ctx, key, storage.KV.GetString)
}

func GetConfigStringMap(ctx context.Context, key string) map[string]any {
	return getc(ctx, key, storage.KV.GetStringMap)
}

func GetConfigStringMapString(ctx context.Context, key string) map[string]string {
	return getc(ctx, key, storage.KV.GetStringMapString)
}

func GetConfigStringMapStringSlice(ctx context.Context, key string) map[string][]string {
	return getc(ctx, key, storage.KV.GetStringMapStringSlice)
}

func GetConfigStringSlice(ctx context.Context, key string) []string {
	return getc(ctx, key, storage.KV.GetStringSlice)
}

func GetConfigTime(ctx context.Context, key string) time.Time {
	return getc(ctx, key, storage.KV.GetTime)
}

func GetConfigUint(ctx context.Context, key string) uint {
	return getc(ctx, key, storage.KV.GetUint)
}

func GetConfigUint16(ctx context.Context, key string) uint16 {
	return getc(ctx, key, storage.KV.GetUint16)
}

func GetConfigUint32(ctx context.Context, key string) uint32 {
	return getc(ctx, key, storage.KV.GetUint32)
}

func GetConfigUint64(ctx context.Context, key string) uint64 {
	return getc(ctx, key, storage.KV.GetUint64)
}

func IsConfigSet(ctx context.Context, key string) bool {
	pctx := contextFrom(ctx)
	return pctx.globalConfig.IsSet(key)
}

func IsSet(ctx context.Context, key string) bool {
	pctx := contextFrom(ctx)
	return pctx.changes.IsSet(key) || pctx.properties.IsSet(key)
}

func Set(ctx context.Context, key string, value any) {
	pctx := contextFrom(ctx)
	pctx.changes.Set(key, value)
}

func get[T any](ctx context.Context, key string, getter func(storage.KV, string) T) T {
	pctx := contextFrom(ctx)
	if pctx.changes.IsSet(key) {
		return getter(pctx.changes, key)
	}
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
