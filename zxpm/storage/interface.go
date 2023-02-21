package storage

import "time"

type KV interface {
	AllSettings() map[string]any
	Get(string) any
	GetBool(string) bool
	GetDuration(string) time.Duration
	GetFloat64(string) float64
	GetInt(string) int
	GetInt32(string) int32
	GetInt64(string) int64
	GetIntSlice(string) []int
	GetString(string) string
	GetStringMap(string) map[string]any
	GetStringMapString(string) map[string]string
	GetStringMapStringSlice(string) map[string][]string
	GetStringSlice(string) []string
	GetTime(string) time.Time
	GetUint(string) uint
	GetUint16(string) uint16
	GetUint32(string) uint32
	GetUint64(string) uint64
	Sub(string) KV

	IsSet(string) bool

	Clear()
	Set(string, any)
	Update(values map[string]any)

	RegisterAlias(string, string)
}

type Requirements interface {
	MarkRequired(string)
	IsRequired(string) bool
	IsPassingRequirements() bool
	MissingRequirements() []string
}
