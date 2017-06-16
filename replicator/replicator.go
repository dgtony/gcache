package replicator

import (
	"github.com/dgtony/gcache/storage"
	"github.com/dgtony/gcache/utils"
	"github.com/op/go-logging"
	//"io"
	"io/ioutil"
	"net"
	"sync"
	"time"
)

const (
	CONN_TIMEOUT          = 10 * time.Second
	CONN_AUTH_WAIT        = 20 * time.Second
	CONN_GET_DUMP_TIMEOUT = 20 * time.Second
	CONN_MAX_IDLE         = 30 * time.Minute
)

var logger *logging.Logger

func init_logger() {
	logger = utils.GetLogger("Replicator")
}

type Replicator struct {
	CacheDump        []byte
	Store            *storage.ConcurrentMap
	DumpFile         string
	MasterAddr       string
	MasterSecretHash []byte
	sync.Mutex
}

func RunReplicator(conf *utils.Config) (*Replicator, *storage.ConcurrentMap) {
	init_logger()

	rep := &Replicator{
		DumpFile:         conf.Replication.CacheFile,
		MasterSecretHash: getSecretHash(conf.Replication.MasterSecret),
		MasterAddr:       conf.Replication.MasterAddr}

	switch conf.Replication.NodeRole {
	case "standalone":
		startStandalone(rep, conf)
	case "master":
		startMaster(rep, conf)
	case "slave":
		startSlave(rep, conf)
	default:
		panic("unsupported node role")
	}

	return rep, rep.Store
}

/* bootstrap procedures */

func startStandalone(rep *Replicator, conf *utils.Config) {
	startStorage(rep, conf)
	if conf.Replication.SaveCacheToFile {
		rep.runDumpUpdater(time.Duration(conf.Replication.DumpUpdatePeriod) * time.Second)
		rep.runFileDumper(time.Duration(conf.Replication.FileWritePeriod) * time.Second)
	}
}

func startMaster(rep *Replicator, conf *utils.Config) {
	startStorage(rep, conf)
	rep.runDumpUpdater(time.Duration(conf.Replication.DumpUpdatePeriod) * time.Second)
	rep.runMasterServer()
	if conf.Replication.SaveCacheToFile {
		rep.runFileDumper(time.Duration(conf.Replication.FileWritePeriod) * time.Second)
	}
}

func startSlave(rep *Replicator, conf *utils.Config) {
	masterConn := startStorageSlave(rep, conf)
	rep.runDumpPuller(masterConn, time.Duration(conf.Replication.DumpUpdatePeriod)*time.Second)
	if conf.Replication.SaveCacheToFile {
		rep.runFileDumper(time.Duration(conf.Replication.FileWritePeriod) * time.Second)
	}
}

func startStorage(rep *Replicator, conf *utils.Config) {
	if conf.Replication.RestoreCacheFromFile {
		dump, err := ioutil.ReadFile(conf.Replication.CacheFile)
		if err == nil {
			store, err := storage.MakeStorageFromDump(conf, dump)
			if err == nil {
				rep.Store = store
				logger.Debug("storage successfully restored from dump")
				return
			}
			logger.Warningf("cannot restore cache from dump: %s", err)
		}
		logger.Warningf("cannot read dump file: %s", err)
	}

	// make empty
	logger.Debug("starting empty cache")
	store, err := storage.MakeStorageEmpty(conf)
	if err != nil {
		panic(err)
	}
	rep.Store = store
}

func startStorageSlave(rep *Replicator, conf *utils.Config) net.Conn {
	conn := ConnectMaster(rep.MasterAddr, CONN_TIMEOUT, rep.MasterSecretHash)
	dump, err := GetMasterDump(conn, CONN_GET_DUMP_TIMEOUT)
	if err != nil {
		panic(err)
	}
	store, err := storage.MakeStorageFromDump(conf, dump)
	if err != nil {
		logger.Errorf("create storage from master snapshot: %s", err)
		panic(err)
	}
	rep.Store = store
	return conn
}

/* replicator proccesses */

// take current snapshot from storage
func (r *Replicator) runDumpUpdater(dumpUpdatePeriod time.Duration) {
	go func() {
		for {
			dump, err := r.Store.DumpStorage()
			if err == nil {
				r.Lock()
				r.CacheDump = dump
				r.Unlock()
			} else {
				logger.Errorf("cannot update cache snapshot: %s", err)
			}
			time.Sleep(dumpUpdatePeriod)
		}
	}()
}

// write snapshot in file
func (r *Replicator) runFileDumper(dumpSavePeriod time.Duration) {
	go func() {
		for {
			time.Sleep(dumpSavePeriod)

			// save current dump in file
			r.Lock()
			data := r.CacheDump
			r.Unlock()
			if err := ioutil.WriteFile(r.DumpFile, data, 0644); err != nil {
				logger.Errorf("cache snapshot saving: %s", err)
			}
		}
	}()
}

// pull storage dump from master (slave only)
func (r *Replicator) runDumpPuller(conn net.Conn, pullDumpPeriod time.Duration) {
	go func() {
		reconFlag := false

		// main loop
		for {
			// try to send request
			dump, err := GetMasterDump(conn, CONN_GET_DUMP_TIMEOUT)
			if err != nil {
				if !reconFlag {
					// try to reconnect
					reconFlag = true
					conn = ConnectMaster(r.MasterAddr, CONN_TIMEOUT, r.MasterSecretHash)
					continue
				} else {
					panic(err)
				}
			}
			reconFlag = false

			// update storage
			if err = r.Store.RestoreFromDump(dump); err != nil {
				logger.Errorf("update storage from master snapshot: %s", err)
				panic(err)
			}

			// update cache dump (for file saving)
			r.Lock()
			r.CacheDump = dump
			r.Unlock()

			time.Sleep(pullDumpPeriod)
		}
	}()
}

// serve storage dump (master only)
func (r *Replicator) runMasterServer() {
	go func() {
		ln, err := net.Listen("tcp", r.MasterAddr)
		if err != nil {
			// no recovery
			panic(err)
		}
		for {
			conn, err := ln.Accept()
			if err != nil {
			}
			go handleSlaveConn(conn, r)
		}
	}()
}
