package plugin

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/zostay/dev-tools/zxpm/plugin/api"
	"github.com/zostay/dev-tools/zxpm/plugin/grpc/client"
	"github.com/zostay/dev-tools/zxpm/plugin/grpc/service"
)

var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "ZXPM_PLUGIN_MAGIC_COOKIE",
	MagicCookieValue: "Q0aHomIRxbv3sa9jlP28A3juUduYTyUnAh4MQnr3",
}

type InterfaceGRPCPlugin struct {
	plugin.Plugin
	Impl Interface
}

func NewPlugin(impl Interface) *InterfaceGRPCPlugin {
	return &InterfaceGRPCPlugin{Impl: impl}
}

func (p *InterfaceGRPCPlugin) GRPCServer(
	_ *plugin.GRPCBroker,
	s *grpc.Server,
) error {
	api.RegisterTaskExecutionServer(s, service.NewGRPCTaskExecution(p.Impl))
	return nil
}

func (p *InterfaceGRPCPlugin) GRPCClient(
	_ context.Context,
	_ *plugin.GRPCBroker,
	c *grpc.ClientConn,
) (any, error) {
	return client.NewGRPCTaskInterface(api.NewTaskExecutionClient(c)), nil
}
