package metal

import (
	goPlugin "github.com/hashicorp/go-plugin"

	"github.com/zostay/dev-tools/zxpm/plugin"
)

func RunPlugin(impl plugin.Interface) {
	goPlugin.Serve(&goPlugin.ServeConfig{
		HandshakeConfig: Handshake,
		Plugins: goPlugin.PluginSet{
			"task-interface": NewPlugin(impl),
		},
		GRPCServer: goPlugin.DefaultGRPCServer,
	})
}
