package main

import (
	"flag"
	"github.com/dgtony/gcache/utils"
	"github.com/op/go-logging"
	// for DEBUG
	//_ "net/http/pprof"
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

	// TODO run replicator (it must internally run core)
	// TODO run API client

}
