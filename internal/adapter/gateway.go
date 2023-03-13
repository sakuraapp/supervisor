package adapter

import (
	gatewaypb "github.com/sakuraapp/protobuf/gateway"
	"github.com/sakuraapp/supervisor/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/resolver"
)

type GatewayAdapter struct {
	gatewaypb.GatewayServiceClient
	conn *grpc.ClientConn
}

func (a *GatewayAdapter) Conn() *grpc.ClientConn {
	return a.conn
}

func NewGatewayAdapter(conf *config.Config) (*GatewayAdapter, error) {
	creds, err := credentials.NewClientTLSFromFile(conf.GatewayKeyPath, "")

	if err != nil {
		return nil, err
	}

	resolver.SetDefaultScheme("dns")

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [ { "round_robin": {} } ] }`), // use round-robin LB
	}

	conn, err := grpc.Dial(conf.GatewayAddr, opts...)

	if err != nil {
		return nil, err
	}

	client := gatewaypb.NewGatewayServiceClient(conn)

	return &GatewayAdapter{
		GatewayServiceClient: client,
		conn: conn,
	}, nil
}