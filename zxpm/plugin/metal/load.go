package metal

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	goPlugin "github.com/hashicorp/go-plugin"

	"github.com/zostay/dev-tools/zxpm/config"
	"github.com/zostay/dev-tools/zxpm/plugin"
)

type Clients map[string]*goPlugin.Client

func buildFirst(cmd string) (string, error) {
	name := path.Base(cmd)
	tmp, err := os.CreateTemp(os.TempDir(), name+"-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary executable file: %w", err)
	}

	err = exec.Command("go", "build", "-o", tmp.Name()).Run()
	if err != nil {
		return "", fmt.Errorf("failed to build %q from %q: %w", tmp.Name(), cmd, err)
	}

	return tmp.Name(), nil
}

const devModePluginPrefix = "go run "

func LoadPlugins(cfg *config.Config) (Clients, error) {
	clients := make(Clients, len(cfg.Plugins))
	for i := range cfg.Plugins {
		pcfg := &cfg.Plugins[i]

		cmd := []string{"sh", "-c", pcfg.Command}
		if strings.HasPrefix(pcfg.Command, devModePluginPrefix) {
			if cfg.Properties.GetBool("DEV_MODE") == false {
				return nil, fmt.Errorf("plugin configuration has plugins in development, but DEV_MODE is not set to true")
			}

			// var err error
			// cmd, err = buildFirst(cmd[len(devModePluginPrefix):])
			// if err != nil {
			// 	return nil, err
			// }
			cmd = []string{"go", "run", pcfg.Command[len(devModePluginPrefix):]}
		}

		// testCmd := exec.Command("go", "run", "github.com/zostay/dev-tools/zxpm/plugin-changelog")
		// testCmd.Stdout = os.Stdout
		// testCmd.Stderr = os.Stderr
		// _ = testCmd.Run()

		client := goPlugin.NewClient(&goPlugin.ClientConfig{
			HandshakeConfig: Handshake,
			Plugins: map[string]goPlugin.Plugin{
				"task-interface": &InterfaceGRPCPlugin{},
			},
			Cmd:              exec.Command(cmd[0], cmd[1:]...),
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
