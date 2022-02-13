package main

import (
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	sharedUtil "github.com/sakuraapp/shared/pkg/util"
	"github.com/sakuraapp/supervisor/internal/config"
	"github.com/sakuraapp/supervisor/internal/service"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		log.WithError(err).Fatal("Failed to load .env file")
	}

	port := os.Getenv("PORT")

	if port == "" {
		port = "9001"
	}

	intPort, err := strconv.ParseInt(port, 10, 64)

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

	conf := &config.Config{
		Env: envType,
		Port: intPort,
		AllowedOrigins: allowedOrigins,
		RedisAddr: redisAddr,
		RedisPassword: redisPassword,
		RedisDatabase: redisDb,
		TLSCertPath: os.Getenv("TLS_CERT_PATH"),
		TLSKeyPath: os.Getenv("TLS_KEY_PATH"),
	}

	_, err = service.New(conf)

	if err != nil {
		log.WithError(err).Fatal("Failed to start service")
	}
}