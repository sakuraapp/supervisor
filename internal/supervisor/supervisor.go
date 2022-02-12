package supervisor

import "github.com/sakuraapp/supervisor/pkg/config"

type Supervisor interface {
	GetConfig() *config.Config
}