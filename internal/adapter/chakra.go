package adapter

import (
	chakrapb "github.com/sakuraapp/protobuf/chakra"
	"github.com/sakuraapp/supervisor/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/resolver"
)

type ChakraAdapter struct {
	chakrapb.ChakraServiceClient
	conn *grpc.ClientConn
}

func (a *ChakraAdapter) Conn() *grpc.ClientConn {
	return a.conn
}

func NewChakraAdapter(conf *config.Config) (*ChakraAdapter, error) {
	creds, err := credentials.NewClientTLSFromFile(conf.ChakraKeyPath, "")

	if err != nil {
		return nil, err
	}

	resolver.SetDefaultScheme("dns")

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [ { "round_robin": {} } ] }`), // use round-robin LB
	}

	conn, err := grpc.Dial(conf.ChakraAddr, opts...)

	if err != nil {
		return nil, err
	}

	client := chakrapb.NewChakraServiceClient(conn)

	return &ChakraAdapter{
		ChakraServiceClient: client,
		conn: conn,
	}, nil
}