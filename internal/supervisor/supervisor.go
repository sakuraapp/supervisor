package supervisor

import (
	"github.com/sakuraapp/supervisor/internal/config"
)

type Supervisor interface {
	GetConfig() *config.Config
}