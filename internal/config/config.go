package config

type envType string

const (
	EnvDEV  envType = "DEV"
	EnvPROD envType = "PROD"
)

type Config struct {
	Env envType
	Port int64
	NodeID string
	AllowedOrigins []string
	RedisAddr string
	RedisPassword string
	RedisDatabase int
	TLSCertPath string
	TLSKeyPath string
}

func (c *Config) IsDev() bool {
	return c.Env == EnvDEV
}