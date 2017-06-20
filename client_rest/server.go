package client_rest

import (
	"github.com/dgtony/gcache/storage"
	"github.com/dgtony/gcache/utils"
	"github.com/op/go-logging"
	"net"
	"net/http"
	"time"
)

var logger *logging.Logger

func StartClientREST(conf *utils.Config, store *storage.ConcurrentMap) error {
	logger = utils.GetLogger("REST")

	serverAddr := net.JoinHostPort(conf.ClientHTTP.Addr, conf.ClientHTTP.Port)
	logger.Infof("client started at %s", serverAddr)
	router := NewRouter(conf, store)

	srv := &http.Server{
		Handler:      router,
		Addr:         serverAddr,
		ReadTimeout:  time.Duration(conf.ClientHTTP.IdleTimeout) * time.Second,
		WriteTimeout: time.Duration(conf.ClientHTTP.IdleTimeout) * time.Second}

	return srv.ListenAndServe()
}
