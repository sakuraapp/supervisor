package service

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	chakrapb "github.com/sakuraapp/protobuf/chakra"
	gatewaypb "github.com/sakuraapp/protobuf/gateway"
	supervisorpb "github.com/sakuraapp/protobuf/supervisor"
	"github.com/sakuraapp/pubsub"
	"github.com/sakuraapp/shared/pkg/model"
	"github.com/sakuraapp/supervisor/internal/adapter"
	"github.com/sakuraapp/supervisor/internal/config"
	"github.com/sakuraapp/supervisor/internal/supervisor"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"net"
	"strconv"
	"time"
)

const maxConnectionAge = 5 * time.Minute

var deployResponse = &supervisorpb.DeployResponse{}
var destroyResponse = &supervisorpb.DestroyResponse{}

type Server struct {
	supervisorpb.UnimplementedSupervisorServiceServer
	pubsub.Dispatcher
	ctx context.Context
	conf *config.Config
	rdb *redis.Client
	queue *Queue
	vmProvider VMProvider
	gateway *adapter.GatewayAdapter
	chakra *adapter.ChakraAdapter
}

func (s *Server) GetConfig() *config.Config {
	return s.conf
}

func (s *Server) GetRedis() *redis.Client {
	return s.rdb
}

func (s *Server) deploy(ctx context.Context, roomId model.RoomId, region supervisor.Region) error {
	strRoomId := strconv.FormatInt(int64(roomId), 10)
	createReq := &chakrapb.CreateRequest{Name: strRoomId}

	stream, err := s.chakra.CreateStream(ctx, createReq)

	if err != nil {
		return err
	}

	currentItemReq := &gatewaypb.SetCurrentItemRequest{
		RoomId: int64(roomId),
		Item: &gatewaypb.CurrentItem{
			Type: gatewaypb.CurrentItem_CHAKRA,
			Url: stream.NodeId,
		},
	}

	_, err = s.gateway.SetCurrentItem(ctx, currentItemReq)

	if err != nil {
		return err
	}

	err s.vmProvider.Deploy(ctx, roomId, region)
}

func (s *Server) Deploy(ctx context.Context, req *supervisorpb.DeployRequest) (*supervisorpb.DeployResponse, error) {
	log.Debugf("Deploy: %+v\n", req)

	roomId := model.RoomId(req.RoomId)
	inQueue, err := s.queue.Has(ctx, roomId)

	if err != nil {
		return nil, err
	}

	if inQueue {
		return deployResponse, nil
	}

	region := supervisor.Region(req.Region.String())
	availableVMs, err := s.vmProvider.GetAvailableRooms(ctx, region)

	if err != nil {
		return nil, err
	}

	if availableVMs == 0 {
		err = s.queue.Add(ctx, roomId)

		if err != nil {
			return nil, err
		}

		return deployResponse, nil
	} else {
		// deploy
		err = s.deploy(ctx, roomId, region)

		if err != nil {
			return nil, err
		}

		return deployResponse, nil
	}
}

func (s *Server) deployNext() error {
	ctx := s.ctx
	roomId, err := s.queue.Pop(ctx)

	if err == redis.Nil {
		return nil
	} else if err != nil {
		return err
	}

	return s.deploy(ctx, roomId)
}

func (s *Server) Destroy(ctx context.Context, req *supervisorpb.DestroyRequest) (*supervisorpb.DestroyResponse, error) {
	log.Debugf("Deploy: %+v\n", req)

	roomId := model.RoomId(req.RoomId)
	err := s.vmProvider.Destroy(ctx, roomId)

	if err != nil {
		return nil, err
	}

	// destroy chakra stream here

	defer func() {
		err := s.deployNext()

		if err != nil {
			log.WithError(err).Error("Failed to deploy the next room in queue")
		}
	}()

	return destroyResponse, nil
}

func New(conf *config.Config) (*Server, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: conf.RedisAddr,
		Password: conf.RedisPassword,
		DB: conf.RedisDatabase,
	})

	vmProvider, err := adapter.NewKubernetesAdapter(conf)

	if err != nil {
		log.WithError(err).Fatal("Failed to instantiate VM Provider")
	}

	g, err := adapter.NewGatewayAdapter(conf)

	if err != nil {
		log.WithError(err).Fatal("Failed to instantiate Gateway Adapter")
	}

	c, err := adapter.NewChakraAdapter(conf)

	if err != nil {
		log.WithError(err).Fatal("Failed to instantiate Chakra Adapter")
	}

	ctx := context.Background()

	d := pubsub.NewRedisDispatcher(ctx, nil, "", rdb)
	s := &Server{
		Dispatcher: d,
		ctx: ctx,
		conf: conf,
		rdb: rdb,
		vmProvider: vmProvider,
		gateway: g,
		chakra: c,
	}

	s.queue = NewQueue(s)

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
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionAge: maxConnectionAge,
		}),
	}

	grpcServer := grpc.NewServer(opts...)
	supervisorpb.RegisterSupervisorServiceServer(grpcServer, s)
	err = grpcServer.Serve(listener)

	if err != nil {
		log.WithError(err).Fatal("Failed to start gRPC server")
	}

	return s, err
}