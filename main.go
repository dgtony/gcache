package main

import (
	"flag"
	"fmt"
	"github.com/dgtony/gcache/client_rest"
	"github.com/dgtony/gcache/replicator"
	"github.com/dgtony/gcache/utils"
	"github.com/op/go-logging"
	// profiling
	//"net/http"
	//_ "net/http/pprof"
)

var logger *logging.Logger

func catch_err() {
	if err := recover(); err != nil {
		fmt.Printf("program error occured => %s\n", err)
	}
}

func main() {
	defer catch_err()

	confFile := flag.String("c", "config.toml", "path to config file")
	flag.Parse()

	// get configuration
	config, err := utils.ReadConfig(*confFile)
	if err != nil {
		panic(err)
	}

	// setup loggers
	utils.SetupLoggers(config)
	logger = utils.GetLogger("Cache")
	logger.Infof("starting cache, node role: %s", config.Replication.NodeRole)

	// run replicator and core storage
	_, store := replicator.RunReplicator(config)

	stopCh := make(chan struct{})
	// run clients
	_ = client_rest.StartClientREST(config, store, stopCh)

	// profiling
	//go http.ListenAndServe("0.0.0.0:7878", nil)

	// wait for all clients to stop
	<-stopCh
	logger.Info("cache server stopped")
}
