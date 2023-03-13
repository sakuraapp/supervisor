package main

import (
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	sharedUtil "github.com/sakuraapp/shared/pkg/util"
	"github.com/sakuraapp/supervisor/internal/config"
	"github.com/sakuraapp/supervisor/internal/service"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/resource"
	"os"
	"strconv"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		log.WithError(err).Fatal("Failed to load .env file")
	}

	strPort := os.Getenv("PORT")
	port, err := strconv.ParseInt(strPort, 10, 64)

	if err != nil {
		log.WithError(err).Fatal("Invalid port")
	}

	env := os.Getenv("APP_ENV")
	envType := config.EnvDEV

	if env == string(config.EnvPROD) {
		envType = config.EnvPROD
	}

	nodeId := os.Getenv("NODE_ID")

	if nodeId == "" {
		nodeId = uuid.NewString()
	}

	allowedOrigins := sharedUtil.ParseAllowedOrigins(os.Getenv("ALLOWED_ORIGINS"))

	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDatabase := os.Getenv("REDIS_DATABASE")
	redisDb, err := strconv.Atoi(redisDatabase)

	if err != nil {
		redisDb = 0
	}

	gatewayAddr := os.Getenv("GATEWAY_ADDR")

	if gatewayAddr == "" {
		log.Fatal("Gateway address is not configured")
	}

	chakraAddr := os.Getenv("CHAKRA_ADDR")

	if chakraAddr == "" {
		log.Fatal("Chakra address is not configured")
	}

	roomImage := os.Getenv("ROOM_IMAGE")

	if roomImage == "" {
		log.Fatal("Invalid room image")
	}

	roomCPULimit, err := resource.ParseQuantity(os.Getenv("ROOM_CPU_LIMIT"))

	if err != nil {
		log.WithError(err).Fatal("Failed to parse room CPU limit")
	}

	roomMemoryLimit, err := resource.ParseQuantity(os.Getenv("ROOM_MEMORY_LIMIT"))

	if err != nil {
		log.WithError(err).Fatal("Failed to parse room memory limit")
	}

	roomCPURequests, err := resource.ParseQuantity(os.Getenv("ROOM_CPU_REQUESTS"))

	if err != nil {
		log.WithError(err).Fatal("Failed to parse room cpu requests")
	}

	roomMemoryRequests, err := resource.ParseQuantity(os.Getenv("ROOM_MEMORY_REQUESTS"))

	if err != nil {
		log.WithError(err).Fatal("Failed to parse room memory requests")
	}

	conf := &config.Config{
		Env: envType,
		Port: int(port),
		AllowedOrigins: allowedOrigins,
		RedisAddr: redisAddr,
		RedisPassword: redisPassword,
		RedisDatabase: redisDb,
		TLSCertPath: os.Getenv("TLS_CERT_PATH"),
		TLSKeyPath: os.Getenv("TLS_KEY_PATH"),
		GatewayAddr: gatewayAddr,
		GatewayKeyPath: os.Getenv("GATEWAY_KEY_PATH"),
		ChakraAddr: chakraAddr,
		ChakraKeyPath: os.Getenv("CHAKRA_KEY_PATH"),
		RoomImage: roomImage,
		RoomCPULimit: roomCPULimit,
		RoomMemoryLimit: roomMemoryLimit,
		RoomCPURequests: roomCPURequests,
		RoomMemoryRequests: roomMemoryRequests,
		K8SConfigPath: os.Getenv("K8S_CONFIG_PATH"),
	}

	_, err = service.New(conf)

	if err != nil {
		log.WithError(err).Fatal("Failed to start service")
	}
}