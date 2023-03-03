package storage

import (
	"strings"
	"time"
)

const ExportPrefix = "__export__."

var _ KV = &KVExp{}

type KVExp struct {
	KV
}

func ExportsOnly(values KV) *KVExp {
	return &KVExp{values}
}

func (e *KVExp) AllKeys() []string {
	keys := e.KV.AllKeys()
	exportKeys := make([]string, 0, len(keys))
	for _, key := range keys {
		if !strings.HasPrefix(key, ExportPrefix) {
			continue
		}
		exportKeys = append(exportKeys, key[len(ExportPrefix):])
	}

	out := make([]string, 0, len(exportKeys))
	for _, key := range exportKeys {
		if !e.KV.IsSet(key) {
			continue
		}
		out = append(out, key)
	}

	return out
}

func (e *KVExp) AllSettings() map[string]any {
	out := New()
	keys := e.AllKeys()
	for _, key := range keys {
		if !e.KV.IsSet(key) {
			continue
		}
		out.Set(key, e.KV.Get(key))
	}
	return out.AllSettings()
}

func (e *KVExp) AllSettingsStrings() map[string]string {
	keys := e.AllKeys()
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		out[key] = e.KV.GetString(key)
	}
	return out
}

func gete[T any](e *KVExp, key string, getter func(KV, string) T) T {
	if e.KV.IsSet(ExportPrefix + key) {
		return getter(e.KV, key)
	}
	return zero[T]()
}

func zero[T any]() T {
	var noop T
	return noop
}

func (e *KVExp) Get(key string) any {
	return gete[any](e, key, KV.Get)
}

func (e *KVExp) GetBool(key string) bool {
	return gete[bool](e, key, KV.GetBool)
}

func (e *KVExp) GetDuration(key string) time.Duration {
	return gete[time.Duration](e, key, KV.GetDuration)
}

func (e *KVExp) GetFloat64(key string) float64 {
	return gete[float64](e, key, KV.GetFloat64)
}

func (e *KVExp) GetInt(key string) int {
	return gete[int](e, key, KV.GetInt)
}

func (e *KVExp) GetInt32(key string) int32 {
	return gete[int32](e, key, KV.GetInt32)
}

func (e *KVExp) GetInt64(key string) int64 {
	return gete[int64](e, key, KV.GetInt64)
}

func (e *KVExp) GetIntSlice(key string) []int {
	return gete[[]int](e, key, KV.GetIntSlice)
}

func (e *KVExp) GetString(key string) string {
	return gete[string](e, key, KV.GetString)
}

func (e *KVExp) GetStringMap(key string) map[string]any {
	return gete[map[string]any](e, key, KV.GetStringMap)
}

func (e *KVExp) GetStringMapString(key string) map[string]string {
	return gete[map[string]string](e, key, KV.GetStringMapString)
}

func (e *KVExp) GetStringMapStringSlice(key string) map[string][]string {
	return gete[map[string][]string](e, key, KV.GetStringMapStringSlice)
}

func (e *KVExp) GetStringSlice(key string) []string {
	return gete[[]string](e, key, KV.GetStringSlice)
}

func (e *KVExp) GetTime(key string) time.Time {
	return gete[time.Time](e, key, KV.GetTime)
}

func (e *KVExp) GetUint(key string) uint {
	return gete[uint](e, key, KV.GetUint)
}

func (e *KVExp) GetUint16(key string) uint16 {
	return gete[uint16](e, key, KV.GetUint16)
}

func (e *KVExp) GetUint32(key string) uint32 {
	return gete[uint32](e, key, KV.GetUint32)
}

func (e *KVExp) GetUint64(key string) uint64 {
	return gete[uint64](e, key, KV.GetUint64)
}

func (e *KVExp) Sub(key string) KV {
	return gete[KV](e, key, KV.Sub)
}

func (e *KVExp) IsSet(key string) bool {
	return gete[bool](e, key, KV.IsSet)
}
