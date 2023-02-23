package plugin

import (
	"github.com/hashicorp/go-plugin"
)

func RunPlugin(impl Interface) {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: Handshake,
		Plugins: plugin.PluginSet{
			"task-interface": NewPlugin(impl),
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
