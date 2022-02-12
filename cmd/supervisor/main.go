package main

import (
	"github.com/joho/godotenv"
	sharedUtil "github.com/sakuraapp/shared/pkg/util"
	"github.com/sakuraapp/supervisor/internal/config"
	"github.com/sakuraapp/supervisor/internal/server"
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

	allowedOrigins := sharedUtil.ParseAllowedOrigins(os.Getenv("ALLOWED_ORIGINS"))

	conf := &config.Config{
		Port: intPort,
		AllowedOrigins: allowedOrigins,
	}

	_, err = server.New(conf)

	if err != nil {
		log.WithError(err).Fatal("Failed to start server")
	}
}