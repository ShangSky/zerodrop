package main

import (
	"flag"

	"github.com/shangsky/zerodrop/config"
	"github.com/shangsky/zerodrop/log"
)

func main() {
	addr := flag.String("addr", ":9889", "server bind address")
	flag.Parse()
	logger := log.New()
	conf := config.New(*addr)
	server := wireAPP(conf)
	logger.Info("server run", "addr", conf.Addr)
	if err := server.ListenAndServe(); err != nil {
		logger.Error("server listen error", "error", err)
	}
	logger.Info("server stoped")
}
