package server

import (
	"fmt"
	"github.com/sakuraapp/supervisor/internal/config"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type Server struct {
	conf *config.Config
}

func (s *Server) GetConfig() *config.Config {
	return s.conf
}

func New(conf *config.Config) (*Server, error) {
	s := &Server{conf: conf}
	r := NewRouter(s)

	log.Printf("Listening on port %v", conf.Port)

	addr := fmt.Sprintf("0.0.0.0:%v", conf.Port)
	err := http.ListenAndServe(addr, r)

	return s, err
}