package service

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	supervisorpb "github.com/sakuraapp/protobuf/supervisor"
	"github.com/sakuraapp/supervisor/internal/config"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net"
)

type Server struct {
	supervisorpb.UnimplementedSupervisorServiceServer
	conf *config.Config
	rdb *redis.Client
}

func (s *Server) GetConfig() *config.Config {
	return s.conf
}

func (s *Server) GetRedis() *redis.Client {
	return s.rdb
}

func (s *Server) Deploy(ctx context.Context, req *supervisorpb.DeployRequest) (*supervisorpb.DeployResponse, error) {
	fmt.Printf("Deploy: %+v\n", req)

	return &supervisorpb.DeployResponse{}, nil
}

func New(conf *config.Config) (*Server, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: conf.RedisAddr,
		Password: conf.RedisPassword,
		DB: conf.RedisDatabase,
	})

	s := &Server{
		conf: conf,
		rdb: rdb,
	}

	creds, err := credentials.NewServerTLSFromFile(conf.TLSCertPath, conf.TLSKeyPath)

	if err != nil {
		log.WithError(err).Fatal("Failed to load SSL/TLS key pair")
	}

	addr := fmt.Sprintf("0.0.0.0:%v", conf.Port)
	listener, err := net.Listen("tcp", addr)

	if err != nil {
		log.WithError(err).Fatal("Failed to start TCP server")
	}

	log.Printf("Listening on port %v", conf.Port)

	opts := []grpc.ServerOption{
		grpc.Creds(creds),
	}

	grpcServer := grpc.NewServer(opts...)
	supervisorpb.RegisterSupervisorServiceServer(grpcServer, s)
	err = grpcServer.Serve(listener)

	if err != nil {
		log.WithError(err).Fatal("Failed to start gRPC server")
	}

	return s, err
}