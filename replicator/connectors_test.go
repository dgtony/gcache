package replicator

import (
	"github.com/dgtony/gcache/utils"
	"testing"
	"time"
)

var logSetFlag bool

func TestReplicatorConnBackoff(t *testing.T) {
	if backoff(0, time.Minute) != 0 {
		t.Error("backoff failure on attempt 0")
	}
	if backoff(1, time.Minute) != time.Second {
		t.Error("backoff failure on attempt 1")
	}
	if backoff(3, time.Minute) != 7*time.Second {
		t.Error("backoff failure on attempt 3")
	}
	if backoff(10, time.Minute) != time.Minute {
		t.Error("backoff failure on attempt 10")
	}
}

func TestReplicatorConnMasterSlave(t *testing.T) {
	defer catch_panic(t)
	setup_logger()

	masterAddr := ":12346"
	secret := "secret"
	cacheDump := []byte("somedatahere")

	// start fake replicator
	rep := Replicator{
		CacheDump:        cacheDump,
		MasterAddr:       masterAddr,
		MasterSecretHash: getSecretHash(secret),
	}
	rep.runMasterServer()

	// connect
	conn := ConnectMaster(masterAddr, 2*time.Second, getSecretHash(secret))

	// get dump
	rcvDump, err := GetMasterDump(conn, 2*time.Second)
	if err != nil {
		t.Errorf("get master dump failure => %s", err)
	}

	if !utils.CompareByteSlices(cacheDump, rcvDump) {
		t.Error("dumps do not match")
	}
}

/* helpers */

func setup_logger() {
	if !logSetFlag {
		utils.SetupLoggers(&utils.Config{
			General: utils.GeneralSettings{
				LogLevel:  "debug",
				LogFormat: "short",
				LogOut:    "stdout"}})
		init_logger()
	}
}

func catch_panic(t *testing.T) {
	if r := recover(); r != nil {
		t.Errorf("panic detected => %s", r)
	}
}
