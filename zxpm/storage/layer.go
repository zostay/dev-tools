package storage

import (
	"sort"
	"time"
)

type KVLayer struct {
	Layers []KV
}

func Layers(layers ...KV) *KVLayer {
	if len(layers) == 0 {
		panic("there must be at least one layer in layered storage")
	}
	nonNilLayers := make([]KV, 0, len(layers))
	for _, layer := range layers {
		if layer == nil {
			continue
		}
		nonNilLayers = append(nonNilLayers, layer)
	}
	return &KVLayer{nonNilLayers}
}

func (l *KVLayer) AllKeys() []string {
	set := make(map[string]struct{}, len(l.Layers[0].AllKeys()))
	for _, layer := range l.Layers {
		for _, k := range layer.AllKeys() {
			set[k] = struct{}{}
		}
	}
	out := keys[struct{}](set)
	sort.Strings(out)
	return out
}

func (l *KVLayer) AllSettings() map[string]any {
	out := make(map[string]any, len(l.Layers[0].AllKeys()))
	for _, l := range l.Layers {
		for _, k := range l.AllKeys() {
			out[k] = l.Get(k)
		}
	}
	return out
}

func (l *KVLayer) AllSettingsStrings() map[string]string {
	keys := l.AllKeys()
	out := make(map[string]string, len(keys))
	for _, k := range keys {
		out[k] = l.GetString(k)
	}
	return out
}

func getl[T any](c *KVLayer, key string, getter func(KV, string) T, zero func() T) T {
	for i := 0; i < len(c.Layers); i++ {
		if c.Layers[i].IsSet(key) {
			return getter(c.Layers[i], key)
		}
	}
	return zero()
}

func (l *KVLayer) Get(key string) any {
	return getl[any](l, key, KV.Get, func() any { return nil })
}

func (l *KVLayer) GetBool(key string) bool {
	return getl[bool](l, key, KV.GetBool, func() bool { return false })
}

func (l *KVLayer) GetDuration(key string) time.Duration {
	return getl[time.Duration](l, key, KV.GetDuration, func() time.Duration { return 0 })
}

func (l *KVLayer) GetFloat64(key string) float64 {
	return getl[float64](l, key, KV.GetFloat64, func() float64 { return 0 })
}

func (l *KVLayer) GetInt(key string) int {
	return getl[int](l, key, KV.GetInt, func() int { return 0 })
}

func (l *KVLayer) GetInt32(key string) int32 {
	return getl[int32](l, key, KV.GetInt32, func() int32 { return 0 })
}

func (l *KVLayer) GetInt64(key string) int64 {
	return getl[int64](l, key, KV.GetInt64, func() int64 { return 0 })
}

func (l *KVLayer) GetIntSlice(key string) []int {
	return getl[[]int](l, key, KV.GetIntSlice, func() []int { return nil })
}

func (l *KVLayer) GetString(key string) string {
	return getl[string](l, key, KV.GetString, func() string { return "" })
}

func (l *KVLayer) GetStringMap(key string) map[string]any {
	return getl[map[string]any](l, key, KV.GetStringMap, func() map[string]any { return nil })
}

func (l *KVLayer) GetStringMapString(key string) map[string]string {
	return getl[map[string]string](l, key, KV.GetStringMapString, func() map[string]string { return nil })
}

func (l *KVLayer) GetStringMapStringSlice(key string) map[string][]string {
	return getl[map[string][]string](l, key, KV.GetStringMapStringSlice, func() map[string][]string { return nil })
}

func (l *KVLayer) GetStringSlice(key string) []string {
	return getl[[]string](l, key, KV.GetStringSlice, func() []string { return nil })
}

func (l *KVLayer) GetTime(key string) time.Time {
	return getl[time.Time](l, key, KV.GetTime, func() time.Time { return time.Time{} })
}

func (l *KVLayer) GetUint(key string) uint {
	return getl[uint](l, key, KV.GetUint, func() uint { return 0 })
}

func (l *KVLayer) GetUint16(key string) uint16 {
	return getl[uint16](l, key, KV.GetUint16, func() uint16 { return 0 })
}

func (l *KVLayer) GetUint32(key string) uint32 {
	return getl[uint32](l, key, KV.GetUint32, func() uint32 { return 0 })
}

func (l *KVLayer) GetUint64(key string) uint64 {
	return getl[uint64](l, key, KV.GetUint64, func() uint64 { return 0 })
}

func (l *KVLayer) Sub(key string) KV {
	newLayers := make([]KV, len(l.Layers))
	for i, layer := range l.Layers {
		newLayers[i] = layer.Sub(key)
	}
	return Layers(newLayers...)
}

func (l *KVLayer) IsSet(key string) bool {
	for _, layer := range l.Layers {
		if layer.IsSet(key) {
			return true
		}
	}
	return false
}

func (l *KVLayer) Clear() {
	for _, layer := range l.Layers {
		layer.Clear()
	}
}

func (l *KVLayer) Set(key string, value any) {
	l.Layers[0].Set(key, value)
}

func (l *KVLayer) Update(values map[string]any) {
	l.Layers[0].Update(values)
}

func (l *KVLayer) UpdateStrings(values map[string]string) {
	l.Layers[0].UpdateStrings(values)
}

func (l *KVLayer) RegisterAlias(alias, key string) {
	for _, layer := range l.Layers {
		layer.RegisterAlias(alias, key)
	}
}
