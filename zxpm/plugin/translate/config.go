package translate

import (
	"github.com/zostay/dev-tools/zxpm/plugin"
	"github.com/zostay/dev-tools/zxpm/plugin/api"
)

func APIConfigToPluginConfig(in *api.Config) *plugin.Config {
	out := &plugin.Config{
		Values:      make(map[string]string, len(in.GetValues())),
		SubSections: make(map[string]*plugin.Config, len(in.GetSubSections())),
	}

	for k, v := range in.GetValues() {
		out.Values[k] = v
	}

	for k, v := range in.GetSubSections() {
		out.SubSections[k] = APIConfigToPluginConfig(v)
	}

	return out
}
