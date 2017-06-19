package storage

import (
	"github.com/dgtony/gcache/utils"
	"testing"
	"time"
)

var logSetFlag bool

func TestCoreCRUDEmpty(t *testing.T) {
	setup_logger()
	if _, err := MakeStorageEmpty(getTestConfig(0)); err == nil {
		t.Error("no error for zero shards")
	}

	if _, err := MakeStorageEmpty(getTestConfig(MAX_SHARDS + 1)); err == nil {
		t.Error("no error for too many shards")
	}

	numShards := 4
	core, err := MakeStorageEmpty(getTestConfig(numShards))
	if err != nil {
		t.Errorf("create empty storage: %s", err)
	}

	storedKeys := core.Keys()
	if len(storedKeys) != 0 {
		t.Errorf("%d keys in empty storage: %+v", len(storedKeys), storedKeys)
	}

	// get nonexistent key
	_, ok := core.Get("non_ex_key")
	if ok {
		t.Error("no indication of nonexistent value")
	}

	// silent remove
	core.Remove("non_ex_key")

}

func TestCoreCRUDValues(t *testing.T) {
	setup_logger()
	numShards := 4
	core, err := MakeStorageEmpty(getTestConfig(numShards))
	if err != nil {
		t.Errorf("create empty storage: %s", err)
	}

	// insert test data
	keyTTL := time.Minute
	testKV := getTestKV()
	for k, v := range testKV {
		if ok := core.Set(k, v, keyTTL); !ok {
			t.Errorf("cannot insert pair %s:%s with ttl %s", k, v, keyTTL)
		}
	}

	// get
	for k, v := range testKV {
		value, ok := core.Get(k)
		if !ok {
			t.Error("cannot get inserted value")
		}
		if !utils.CompareByteSlices(v, value) {
			t.Error("inserted value corrupted")
		}
	}

	// remove
	keyToRemove := "key3"
	core.Remove(keyToRemove)
	if _, ok := core.Get(keyToRemove); ok {
		t.Error("removed key still available")
	}

	// keys
	storedKeys := core.Keys()
	if len(storedKeys) != len(testKV)-1 {
		t.Errorf("expected keys: %d, found in storage: %d", len(storedKeys)-1, len(storedKeys))
	}

	// look for removed in stored
	if utils.FindInSliceString(keyToRemove, storedKeys) {
		t.Error("removed key still returned in keys")
	}
}

func TestCoreIntegrationDumpRestore(t *testing.T) {
	setup_logger()
	numShards := 4
	core, err := MakeStorageEmpty(getTestConfig(numShards))
	if err != nil {
		t.Errorf("create empty storage: %s", err)
	}

	keyTTL := 1 * time.Minute
	testKV := getTestKV()
	for k, v := range testKV {
		if ok := core.Set(k, v, keyTTL); !ok {
			t.Errorf("cannot insert pair %s:%s with ttl %s", k, v, keyTTL)
		}
	}

	// make dump
	dump, err := core.DumpStorage()

	// remove keys from original storage
	for k, _ := range testKV {
		core.Remove(k)
	}
	if len(core.Keys()) > 0 {
		t.Error("some keys weren't removed")
	}

	// try to restore from dump
	err = core.RestoreFromDump(dump)
	if err != nil {
		t.Errorf("cannot restore storage from dump: %s", err)
	}

	// wait until entire storage will be restored
	time.Sleep(50 * time.Millisecond)

	// verify
	for k, v := range testKV {
		value, ok := core.Get(k)
		if !ok || !utils.CompareByteSlices(v, value) {
			t.Error("original and restored storages doesn't match")
		}
	}
}

/* helpers */

func getTestConfig(numShards int) *utils.Config {
	return &utils.Config{
		General: utils.GeneralSettings{
			LogLevel:  "debug",
			LogFormat: "short",
			LogOut:    "stdout"},
		Replication: utils.ReplicationSettings{},
		Storage: utils.StorageSettings{
			NumShards:               numShards,
			ExpiredKeyCheckInterval: 60}}
}

func getTestKV() map[string][]byte {
	return map[string][]byte{
		"key1": []byte("value1"),
		"key2": []byte("value2"),
		"key3": []byte("value3"),
		"key4": []byte("value4"),
		"key5": []byte("value5")}
}

func setup_logger() {
	if !logSetFlag {
		logConf := getTestConfig(1)
		utils.SetupLoggers(logConf)
	}
}
