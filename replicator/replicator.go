package replicator

import (
	"github.com/dgtony/gcache/storage"
)

type Dump struct{}

type Replicator struct {
	CacheDump Dump
	// ?
	Store *storage.ConcurrentMap
}

func startStandalone() {

}

func startMaster() {

}

func startSlave() {

}
