package config

import "k8s.io/apimachinery/pkg/api/resource"

type envType string

const (
	EnvDEV  envType = "DEV"
	EnvPROD envType = "PROD"
)

type Config struct {
	Env envType
	Port int
	NodeID string
	AllowedOrigins []string
	RedisAddr string
	RedisPassword string
	RedisDatabase int
	TLSCertPath string
	TLSKeyPath string
	GatewayAddr string
	GatewayKeyPath string
	ChakraAddr string
	ChakraKeyPath string
	RoomImage string
	RoomCPULimit resource.Quantity
	RoomMemoryLimit resource.Quantity
	RoomCPURequests resource.Quantity
	RoomMemoryRequests resource.Quantity
	K8SConfigPath string
}

func (c *Config) IsDev() bool {
	return c.Env == EnvDEV
}