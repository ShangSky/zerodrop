//go:build wireinject

package main

import (
	"net/http"

	"github.com/google/wire"
	"github.com/shangsky/zerodrop/config"
	"github.com/shangsky/zerodrop/handler"
	"github.com/shangsky/zerodrop/log"
	"github.com/shangsky/zerodrop/server"
)

func wireAPP(conf *config.Config) *http.Server {
	wire.Build(
		log.ProviderSet,
		handler.ProviderSet,
		server.ProviderSet,
	)
	return nil
}
