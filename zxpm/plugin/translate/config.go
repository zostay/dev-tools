package translate

import (
	"github.com/zostay/dev-tools/zxpm/plugin/api"
	"github.com/zostay/dev-tools/zxpm/storage"
)

func APIConfigToKV(in *api.Config) *storage.KVMem {
	out := storage.New()

	for k, v := range in.GetValues() {
		out.Set(k, v)
	}

	return out
}

func KVToAPIConfig(in storage.KV) *api.Config {
	return &api.Config{Values: KVToStringMapString(in)}
}

func KVToStringMapString(in storage.KV) map[string]string {
	keys := in.AllKeys()
	out := make(map[string]string, len(keys))

	for _, k := range keys {
		out[k] = in.GetString(k)
	}
	return out
}
