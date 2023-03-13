package supervisor

import (
	"github.com/go-redis/redis/v8"
	"github.com/sakuraapp/supervisor/internal/config"
)

const (
	RegionANY = "any"
	RegionEUW = "euw"
)

type Region string

type App interface {
	GetConfig() *config.Config
	GetRedis() *redis.Client
}