package storage

import (
	"sort"
	"time"
)

var _ KV = &KVChanges{}

type KVChanges struct {
	changes KV
	Inner   KV
}

func WithChangeTracking(inner KV) *KVChanges {
	return &KVChanges{
		Inner:   inner,
		changes: New(),
	}
}

func (c *KVChanges) AllKeys() []string {
	var (
		innerKeys   = c.Inner.AllKeys()
		changesKeys = c.changes.AllKeys()
		set         = make(map[string]struct{}, len(innerKeys)+len(changesKeys))
	)
	for _, k := range innerKeys {
		set[k] = struct{}{}
	}
	for _, k := range changesKeys {
		set[k] = struct{}{}
	}
	out := keys[struct{}](set)
	sort.Strings(out)
	return out
}

func (c *KVChanges) AllSettings() map[string]any {
	var (
		innerKeys   = c.Inner.AllKeys()
		changesKeys = c.changes.AllKeys()
		out         = make(map[string]any, len(innerKeys)+len(changesKeys))
	)
	for _, k := range innerKeys {
		out[k] = c.Inner.Get(k)
	}
	for _, k := range changesKeys {
		out[k] = c.changes.Get(k)
	}
	return out
}

func get[T any](c *KVChanges, key string, getter func(KV, string) T) T {
	if c.changes.IsSet(key) {
		return getter(c.changes, key)
	}
	return getter(c.Inner, key)
}

func (c *KVChanges) Get(key string) any {
	return get[any](c, key, KV.Get)
}

func (c *KVChanges) GetBool(key string) bool {
	return get[bool](c, key, KV.GetBool)
}

func (c *KVChanges) GetDuration(key string) time.Duration {
	return get[time.Duration](c, key, KV.GetDuration)
}

func (c *KVChanges) GetFloat64(key string) float64 {
	return get[float64](c, key, KV.GetFloat64)
}

func (c *KVChanges) GetInt(key string) int {
	return get[int](c, key, KV.GetInt)
}

func (c *KVChanges) GetInt32(key string) int32 {
	return get[int32](c, key, KV.GetInt32)
}

func (c *KVChanges) GetInt64(key string) int64 {
	return get[int64](c, key, KV.GetInt64)
}

func (c *KVChanges) GetIntSlice(key string) []int {
	return get[[]int](c, key, KV.GetIntSlice)
}

func (c *KVChanges) GetString(key string) string {
	return get[string](c, key, KV.GetString)
}

func (c *KVChanges) GetStringMap(key string) map[string]any {
	return get[map[string]any](c, key, KV.GetStringMap)
}

func (c *KVChanges) GetStringMapString(key string) map[string]string {
	return get[map[string]string](c, key, KV.GetStringMapString)
}

func (c *KVChanges) GetStringMapStringSlice(key string) map[string][]string {
	return get[map[string][]string](c, key, KV.GetStringMapStringSlice)
}

func (c *KVChanges) GetStringSlice(key string) []string {
	return get[[]string](c, key, KV.GetStringSlice)
}

func (c *KVChanges) GetTime(key string) time.Time {
	return get[time.Time](c, key, KV.GetTime)
}

func (c *KVChanges) GetUint(key string) uint {
	return get[uint](c, key, KV.GetUint)
}

func (c *KVChanges) GetUint16(key string) uint16 {
	return get[uint16](c, key, KV.GetUint16)
}

func (c *KVChanges) GetUint32(key string) uint32 {
	return get[uint32](c, key, KV.GetUint32)
}

func (c *KVChanges) GetUint64(key string) uint64 {
	return get[uint64](c, key, KV.GetUint64)
}

func (c *KVChanges) Sub(key string) KV {
	return &KVChanges{
		changes: c.changes.Sub(key),
		Inner:   c.Inner.Sub(key),
	}
}

func (c *KVChanges) IsSet(key string) bool {
	return c.changes.IsSet(key) || c.Inner.IsSet(key)
}

func (c *KVChanges) Clear() {
	c.changes.Clear()
	c.Inner.Clear()
}

func (c *KVChanges) Set(key string, value any) {
	c.changes.Set(key, value)
}

func (c *KVChanges) Update(values map[string]any) {
	c.changes.Update(values)
}

func (c *KVChanges) RegisterAlias(alias, key string) {
	c.changes.RegisterAlias(alias, key)
	c.Inner.RegisterAlias(alias, key)
}

func (c *KVChanges) Changes() map[string]any {
	return c.changes.AllSettings()
}

func (c *KVChanges) ChangesFlattenedToString() map[string]string {
	changesKeys := c.changes.AllKeys()
	out := make(map[string]string, len(changesKeys))
	for _, key := range changesKeys {
		out[key] = c.changes.GetString(key)
	}
	return out
}

func (c *KVChanges) ClearChanges() {
	c.changes.Clear()
}