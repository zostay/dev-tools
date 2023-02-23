package plugin

import (
	"fmt"
	"os/exec"

	"github.com/hashicorp/go-plugin"

	"github.com/zostay/dev-tools/zxpm/config"
)

type Clients map[string]*plugin.Client

func LoadPlugins(cfg *config.Config) Clients {
	clients := make(Clients, len(cfg.Plugins))
	for i := range cfg.Plugins {
		pcfg := &cfg.Plugins[i]
		client := plugin.NewClient(&plugin.ClientConfig{
			HandshakeConfig: Handshake,
			Plugins: map[string]plugin.Plugin{
				"task-interface": &InterfaceGRPCPlugin{},
			},
			Cmd:              exec.Command("sh", "-c", pcfg.Command),
			AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		})

		clients[pcfg.Name] = client
	}
	return clients
}

func Dispense(clients Clients, name string) (Interface, error) {
	client, err := clients[name].Client()
	if err != nil {
		return nil, fmt.Errorf("error connecting to plugin %q: %w", name, err)
	}

	raw, err := client.Dispense("task-interface")
	if err != nil {
		return nil, fmt.Errorf("error dispensing plugin %q: %w", name, err)
	}

	iface := raw.(Interface)
	return iface, nil
}

func DispenseAll(clients Clients) (map[string]Interface, error) {
	ifaces := make(map[string]Interface, len(clients))
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
