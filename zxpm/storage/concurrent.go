package storage

import (
	"sync"
	"time"
)

var _ KV = &KVCon{}

type KVCon struct {
	inner KV
	lock  *sync.RWMutex
}

func WithSafeConcurrency(inner KV) *KVCon {
	return &KVCon{inner, &sync.RWMutex{}}
}

func readSyncUU[Out any](
	c *KVCon,
	call func(KV) Out,
) Out {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return call(c.inner)
}

func readSyncBU[In, Out any](
	c *KVCon,
	in In,
	call func(KV, In) Out,
) Out {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return call(c.inner, in)
}

func (c *KVCon) AllKeys() []string {
	return readSyncUU[[]string](c, KV.AllKeys)
}

func (c *KVCon) AllSettings() map[string]any {
	return readSyncUU[map[string]any](c, KV.AllSettings)
}

func (c *KVCon) AllSettingsStrings() map[string]string {
	return readSyncUU[map[string]string](c, KV.AllSettingsStrings)
}

func (c *KVCon) Get(key string) any {
	return readSyncBU[string, any](c, key, KV.Get)
}

func (c *KVCon) GetBool(key string) bool {
	return readSyncBU[string, bool](c, key, KV.GetBool)
}

func (c *KVCon) GetDuration(key string) time.Duration {
	return readSyncBU[string, time.Duration](c, key, KV.GetDuration)
}

func (c *KVCon) GetFloat64(key string) float64 {
	return readSyncBU[string, float64](c, key, KV.GetFloat64)
}

func (c *KVCon) GetInt(key string) int {
	return readSyncBU[string, int](c, key, KV.GetInt)
}

func (c *KVCon) GetInt32(key string) int32 {
	return readSyncBU[string, int32](c, key, KV.GetInt32)
}

func (c *KVCon) GetInt64(key string) int64 {
	return readSyncBU[string, int64](c, key, KV.GetInt64)
}

func (c *KVCon) GetIntSlice(key string) []int {
	return readSyncBU[string, []int](c, key, KV.GetIntSlice)
}

func (c *KVCon) GetString(key string) string {
	return readSyncBU[string, string](c, key, KV.GetString)
}

func (c *KVCon) GetStringMap(key string) map[string]any {
	return readSyncBU[string, map[string]any](c, key, KV.GetStringMap)
}

func (c *KVCon) GetStringMapString(key string) map[string]string {
	return readSyncBU[string, map[string]string](c, key, KV.GetStringMapString)
}

func (c *KVCon) GetStringMapStringSlice(key string) map[string][]string {
	return readSyncBU[string, map[string][]string](c, key, KV.GetStringMapStringSlice)
}

func (c *KVCon) GetStringSlice(key string) []string {
	return readSyncBU[string, []string](c, key, KV.GetStringSlice)
}

func (c *KVCon) GetTime(key string) time.Time {
	return readSyncBU[string, time.Time](c, key, KV.GetTime)
}

func (c *KVCon) GetUint(key string) uint {
	return readSyncBU[string, uint](c, key, KV.GetUint)
}

func (c *KVCon) GetUint16(key string) uint16 {
	return readSyncBU[string, uint16](c, key, KV.GetUint16)
}

func (c *KVCon) GetUint32(key string) uint32 {
	return readSyncBU[string, uint32](c, key, KV.GetUint32)
}

func (c *KVCon) GetUint64(key string) uint64 {
	return readSyncBU[string, uint64](c, key, KV.GetUint64)
}

func (c *KVCon) Sub(key string) KV {
	return &KVCon{c.inner.Sub(key), c.lock}
}

func (c *KVCon) IsSet(key string) bool {
	return readSyncBU[string, bool](c, key, KV.IsSet)
}

func (c *KVCon) Clear() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.inner.Clear()
}

func (c *KVCon) Set(key string, value any) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.inner.Set(key, value)
}

func (c *KVCon) Update(values map[string]any) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.inner.Update(values)
}

func (c *KVCon) UpdateStrings(values map[string]string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.inner.UpdateStrings(values)
}

func (c *KVCon) RegisterAlias(alias, key string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.inner.RegisterAlias(alias, key)
}

func (c *KVCon) Atomic(call func(KV)) {
	c.lock.Lock()
	defer c.lock.Unlock()
	call(c.inner)
}
