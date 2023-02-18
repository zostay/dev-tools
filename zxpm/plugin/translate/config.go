package translate

import (
	"github.com/zostay/dev-tools/pkg/config"
	"github.com/zostay/dev-tools/zxpm/plugin/api"
)

func APIConfigToConfig(in *api.Config) *config.Config {
	out := config.NewCap(len(in.GetValues()), len(in.GetSubSections()))

	for k, v := range in.GetValues() {
		out.Values[k] = v
	}

	for k, v := range in.GetSubSections() {
		out.SubSections[k] = APIConfigToConfig(v)
	}

	return out
}

func ConfigToAPIConfig(in *config.Config) *api.Config {
	out := &api.Config{
		Values:      make(map[string]string, len(in.Values)),
		SubSections: make(map[string]*api.Config, len(in.SubSections)),
	}

	for k, v := range in.Values {
		out.Values[k] = v
	}

	for k, v := range in.SubSections {
		out.SubSections[k] = ConfigToAPIConfig(v)
	}

	return out
}
