package controller

import (
	"github.com/sakuraapp/supervisor/internal/supervisor"
	"net/http"
)

type VMController struct {
	app supervisor.Supervisor
}

func (c *VMController) Deploy(w http.ResponseWriter, r *http.Request) {

}

func Init(app supervisor.Supervisor) *VMController {
	return &VMController{app: app}
}