package metal

import (
	"fmt"
	"os/exec"

	goPlugin "github.com/hashicorp/go-plugin"

	"github.com/zostay/dev-tools/zxpm/config"
	"github.com/zostay/dev-tools/zxpm/plugin"
)

type Clients map[string]*goPlugin.Client

func LoadPlugins(cfg *config.Config) Clients {
	clients := make(Clients, len(cfg.Plugins))
	for i := range cfg.Plugins {
		pcfg := &cfg.Plugins[i]
		client := goPlugin.NewClient(&goPlugin.ClientConfig{
			HandshakeConfig: Handshake,
			Plugins: map[string]goPlugin.Plugin{
				"task-interface": &InterfaceGRPCPlugin{},
			},
			Cmd:              exec.Command("sh", "-c", pcfg.Command),
			AllowedProtocols: []goPlugin.Protocol{goPlugin.ProtocolGRPC},
		})

		clients[pcfg.Name] = client
	}
	return clients
}

func Dispense(clients Clients, name string) (plugin.Interface, error) {
	client, err := clients[name].Client()
	if err != nil {
		return nil, fmt.Errorf("error connecting to plugin %q: %w", name, err)
	}

	raw, err := client.Dispense("task-interface")
	if err != nil {
		return nil, fmt.Errorf("error dispensing plugin %q: %w", name, err)
	}

	iface := raw.(plugin.Interface)
	return iface, nil
}

func DispenseAll(clients Clients) (map[string]plugin.Interface, error) {
	ifaces := make(map[string]plugin.Interface, len(clients))
	for name := range clients {
		iface, err := Dispense(clients, name)
		if err != nil {
			return nil, err
		}

		ifaces[name] = iface
	}
	return ifaces, nil
}

func KillPlugins(clients Clients) {
	for _, v := range clients {
		v.Kill()
	}
}
