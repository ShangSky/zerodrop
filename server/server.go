package server

import (
	"net/http"

	"github.com/shangsky/zerodrop/config"
)

func New(
	conf *config.Config,
	handler http.Handler,
) *http.Server {
	return &http.Server{Addr: conf.Addr, Handler: handler}
}
