package storage

import (
	"sort"
	"strings"

	"github.com/spf13/viper"
)

var (
	_ KV           = &KVViper{}
	_ Requirements = &KVViper{}
)

type KVViper struct {
	*viper.Viper
	requirements map[string]struct{}
}

func NewViper(v *viper.Viper) *KVViper {
	if v == nil {
		v = viper.GetViper()
	}
	return &KVViper{Viper: v}
}

func (v *KVViper) AllSettingsStrings() map[string]string {
	keys := v.AllKeys()
	out := make(map[string]string, len(keys))
	for _, k := range keys {
		out[k] = v.GetString(k)
	}
	return out
}

func (v *KVViper) Clear() {
	v.Viper = viper.New()
}

func (v *KVViper) Update(values map[string]any) {
	for k, val := range values {
		v.Set(k, val)
	}
}

func (v *KVViper) UpdateStrings(values map[string]string) {
	for k, val := range values {
		v.Set(k, val)
	}
}

func (v *KVViper) Sub(key string) KV {
	return &KVViper{v.Viper.Sub(key), v.subRequirements(key)}
}

func (v *KVViper) subRequirements(key string) map[string]struct{} {
	prefix := strings.ToLower(key) + "."
	res := make(map[string]struct{}, len(v.requirements))
	for k := range v.requirements {
		if strings.HasPrefix(k, prefix) {
			subK := k[len(prefix):]
			res[subK] = struct{}{}
		}
	}
	return res
}

func (v *KVViper) MarkRequired(key string) {
	key = strings.ToLower(key)
	v.requirements[key] = struct{}{}
}

func (v *KVViper) IsRequired(key string) bool {
	key = strings.ToLower(key)
	_, exists := v.requirements[key]
	return exists
}

func (v *KVViper) IsPassingRequirements() bool {
	for k := range v.requirements {
		if !v.IsSet(k) {
			return false
		}
	}
	return true
}

func (v *KVViper) MissingRequirements() []string {
	out := make([]string, 0, len(v.requirements))
	for k := range v.requirements {
		if !v.IsSet(k) {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}
