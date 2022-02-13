package supervisor

import (
	"github.com/go-redis/redis/v8"
	"github.com/sakuraapp/supervisor/internal/config"
)

type App interface {
	GetConfig() *config.Config
	GetRedis() *redis.Client
}