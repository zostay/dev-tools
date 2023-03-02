package storage

import (
	"container/list"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cast"
)

var (
	_ KV           = &KVMem{}
	_ Requirements = &KVMem{}
)

type KVMem struct {
	prefix       string
	values       map[string]any
	requirements map[string]struct{}
	aliases      map[string]string
}

func New() *KVMem {
	return &KVMem{
		values:       make(map[string]any, 10),
		requirements: make(map[string]struct{}, 10),
	}
}

func (m *KVMem) RO() *KVCfg {
	return &KVCfg{*m}
}

func (m *KVMem) prefixKey(key string) string {
	if m.prefix != "" {
		return m.prefix + "." + key
	}
	return key
}

func (m *KVMem) resolveAlias(key string) string {
	prefixedKey := m.prefixKey(key)
	if otherKey, isAliased := m.aliases[prefixedKey]; isAliased {
		return m.prefixKey(otherKey)
	}
	return prefixedKey
}

func (m *KVMem) key(key string) string {
	key = strings.ToLower(key)
	return m.resolveAlias(key)
}

func (m *KVMem) splitKey(key string) []string {
	return strings.Split(m.key(key), ".")
}

func (m *KVMem) getExists(key string) (any, bool) {
	keys := m.splitKey(key)
	n := m.values
	for i, k := range keys {
		if n == nil {
			return n, false
		}

		val, exists := n[k]
		if !exists {
			return nil, false
		} else if i == len(keys)-1 {
			// last key being resolved, return the value
			return val, exists
		}

		// not the last key, so we need map[string]any
		switch v := val.(type) {
		case map[string]any:
			n = v
		default:
			// we can't go further, but they want us to, so the key they
			// want doesn't exist
			return nil, false
		}
	}
	return n, true
}

func (m *KVMem) get(key string) any {
	v, _ := m.getExists(key)
	return v
}

func (m *KVMem) set(key string, value any) {
	keys := m.splitKey(key)
	var lastKey string
	keys, lastKey = keys[:len(keys)-1], keys[len(keys)-1]
	values := m.values
	for _, k := range keys {
		xVals := values[k]
		if vals, isStringMap := xVals.(map[string]any); isStringMap {
			values = vals
		} else {
			vals := map[string]any{}
			values, values[k] = vals, vals
		}
	}
	values[lastKey] = value
}

func keys[T any](v map[string]T) []string {
	ks := make([]string, 0, len(v))
	for k := range v {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

func (m *KVMem) AllKeys() []string {
	type openItem struct {
		prefix string
		keys   []string
		in     map[string]any
	}

	out := make([]string, 0, len(m.values))
	openList := list.New()
	openList.PushFront(&openItem{
		prefix: "",
		keys:   keys[any](m.values),
		in:     m.values,
	})
	for openList.Len() > 0 {
		el := openList.Front()
		item := el.Value.(*openItem)

		if len(item.keys) == 0 {
			openList.Remove(el)
			continue
		}

		var key string
		key, item.keys = item.keys[0], item.keys[1:]
		value := item.in[key]

		if item.prefix != "" {
			key = item.prefix + "." + key
		}

		if nextIn, isStringMap := value.(map[string]any); isStringMap {
			openList.PushFront(&openItem{
				prefix: key,
				keys:   keys[any](nextIn),
				in:     nextIn,
			})
			continue
		}

		// only append the key name if it's a value, not a nested map[string]any
		out = append(out, key)
	}

	out = append(out, keys[string](m.aliases)...)
	sort.Strings(out)

	return out
}

func (m *KVMem) AllSettings() map[string]any {
	out := make(map[string]any, len(m.values))
	for k, v := range m.values {
		out[k] = v
	}

	tmp := &KVMem{values: out}
	for alias, key := range m.aliases {
		tmp.set(alias, tmp.get(key))
	}

	return tmp.values
}

func (m *KVMem) AllSettingsStrings() map[string]string {
	keys := m.AllKeys()
	out := make(map[string]string, len(keys))
	for _, k := range keys {
		out[k] = m.GetString(k)
	}
	return out
}

func (m *KVMem) Get(key string) any {
	return m.get(key)
}

func (m *KVMem) GetBool(key string) bool {
	return cast.ToBool(m.get(key))
}

func (m *KVMem) GetDuration(key string) time.Duration {
	return cast.ToDuration(m.get(key))
}

func (m *KVMem) GetFloat64(key string) float64 {
	return cast.ToFloat64(m.get(key))
}

func (m *KVMem) GetInt(key string) int {
	return cast.ToInt(m.get(key))
}

func (m *KVMem) GetInt32(key string) int32 {
	return cast.ToInt32(m.get(key))
}

func (m *KVMem) GetInt64(key string) int64 {
	return cast.ToInt64(m.get(key))
}

func (m *KVMem) GetIntSlice(key string) []int {
	return cast.ToIntSlice(m.get(key))
}

func (m *KVMem) GetString(key string) string {
	return cast.ToString(m.get(key))
}

func (m *KVMem) GetStringMap(key string) map[string]any {
	return cast.ToStringMap(m.get(key))
}

func (m *KVMem) GetStringMapString(key string) map[string]string {
	return cast.ToStringMapString(m.get(key))
}

func (m *KVMem) GetStringMapStringSlice(key string) map[string][]string {
	return cast.ToStringMapStringSlice(m.get(key))
}

func (m *KVMem) GetStringSlice(key string) []string {
	return cast.ToStringSlice(m.get(key))
}

func (m *KVMem) GetTime(key string) time.Time {
	return cast.ToTime(m.get(key))
}

func (m *KVMem) GetUint(key string) uint {
	return cast.ToUint(m.get(key))
}

func (m *KVMem) GetUint16(key string) uint16 {
	return cast.ToUint16(m.get(key))
}

func (m *KVMem) GetUint32(key string) uint32 {
	return cast.ToUint32(m.get(key))
}

func (m *KVMem) GetUint64(key string) uint64 {
	return cast.ToUint64(m.get(key))
}

func (m *KVMem) Sub(key string) KV {
	v := m.get(key)
	if _, isStringMap := v.(map[string]any); isStringMap {
		return &KVMem{
			prefix:       key,
			values:       m.values,
			requirements: m.requirements,
			aliases:      m.aliases,
		}
	}
	return nil
}

func (m *KVMem) IsSet(key string) bool {
	_, exists := m.getExists(key)
	return exists
}

func (m *KVMem) Clear() {
	m.values = make(map[string]any, 10)
}

func (m *KVMem) Set(key string, value any) {
	m.set(key, value)
}

func (m *KVMem) Update(values map[string]any) {
	for k, v := range values {
		m.set(k, v)
	}
}

func (m *KVMem) UpdateStrings(values map[string]string) {
	for k, v := range values {
		m.set(k, v)
	}
}

func (m *KVMem) RegisterAlias(alias, key string) {
	m.aliases[m.key(alias)] = m.key(key)
}

func (m *KVMem) MarkRequired(key string) {
	m.requirements[m.key(key)] = struct{}{}
}

func (m *KVMem) IsRequired(key string) bool {
	_, exists := m.requirements[m.key(key)]
	return exists
}

func (m *KVMem) IsPassingRequirements() bool {
	for k := range m.requirements {
		if !m.IsSet(k) {
			return false
		}
	}
	return true
}

func (m *KVMem) MissingRequirements() []string {
	missing := make([]string, 0, len(m.requirements))
	for k := range m.requirements {
		if !m.IsSet(k) {
			missing = append(missing, k)
		}
	}
	sort.Strings(missing)
	return missing
}
