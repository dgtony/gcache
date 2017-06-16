package main

import (
	"flag"
	"github.com/dgtony/gcache/replicator"
	"github.com/dgtony/gcache/storage"
	"github.com/dgtony/gcache/utils"
	"github.com/op/go-logging"

	// for DEBUG
	//"net/http"
	//_ "net/http/pprof"
	"sync"
	"time"
)

var logger *logging.Logger

func main() {
	confFile := flag.String("c", "config.toml", "path to config file")
	flag.Parse()

	// read config
	config, err := utils.ReadConfig(*confFile)
	if err != nil {
		panic(err)
	}

	// setup loggers
	utils.SetupLoggers(config)
	logger = utils.GetLogger("Cache")

	// TODO remove
	logger.Debugf("config: %+v", config)

	// TODO run replicator (it must internally run core)
	rep, store := replicator.RunReplicator(config)

	// TODO run API client

	// TODO remove
	//go panic(http.ListenAndServe(":8080", nil))
	if config.Replication.NodeRole == "master" {
		store.Set("testkey", []byte("testvalue"), time.Minute)
	}
	stub_wait(rep, store)

}

func stub_wait(rep *replicator.Replicator, s *storage.ConcurrentMap) {
	var wg sync.WaitGroup
	wg.Add(1)
	logger.Debug("client imitation...")
	logger.Debugf("master hash: %x", rep.MasterSecretHash)
	logger.Debugf("keys in storage: %s", s.Keys())
	wg.Wait()
}
