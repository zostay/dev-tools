package metal

import (
	"fmt"
	"os/exec"
	"strings"

	goPlugin "github.com/hashicorp/go-plugin"

	"github.com/zostay/dev-tools/zxpm/config"
	"github.com/zostay/dev-tools/zxpm/plugin"
)

type Clients map[string]*goPlugin.Client

const devModePluginPrefix = "go run "

func LoadPlugins(cfg *config.Config) (Clients, error) {
	clients := make(Clients, len(cfg.Plugins))
	for i := range cfg.Plugins {
		pcfg := &cfg.Plugins[i]

		cmd := []string{"sh", "-c", pcfg.Command}
		if strings.HasPrefix(pcfg.Command, devModePluginPrefix) {
			if !cfg.Properties.GetBool("DEV_MODE") {
				return nil, fmt.Errorf("plugin configuration has plugins in development, but DEV_MODE is not set to true")
			}

			cmd = []string{"go", "run", pcfg.Command[len(devModePluginPrefix):]}
		}

		client := goPlugin.NewClient(&goPlugin.ClientConfig{
			HandshakeConfig: Handshake,
			Plugins: map[string]goPlugin.Plugin{
				"task-interface": &InterfaceGRPCPlugin{},
			},
			Cmd:              exec.Command(cmd[0], cmd[1:]...), //nolint:gosec // foot guns have been handed to user, so tainted value here is expected
			AllowedProtocols: []goPlugin.Protocol{goPlugin.ProtocolGRPC},
		})

		clients[pcfg.Name] = client
	}
	return clients, nil
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
