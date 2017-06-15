package core

import (
	"errors"
	//"fmt"
	"github.com/dgtony/gcache/utils"
	"sync"
	"time"
)

const (
	MAX_SHARDS = 4096
	// key length limit
	KEY_MAX_LEN = 2048
	// value size limit up to 10Mb
	VALUE_MAX_SIZE = 10485760
	// period of key expiration check, sec
	//KEY_EXPIRATION_CHECK_INTERVAL = 30

)

type ConcurrentMap []*ConcurrentMapShard

type ConcurrentMapShard struct {
	Items         map[string][]byte
	KeyExpiration ExpireQueue
	sync.RWMutex
}

/* Storage methods */

// TODO run procedures on shard ??

func MakeStorageEmpty(numShards int) (ConcurrentMap, error) {
	if numShards < 1 || numShards > MAX_SHARDS {
		return nil, errors.New("wrong number of shards")
	}

	m := make(ConcurrentMap, numShards)
	for i := 0; i < numShards; i++ {
		m[i] = &ConcurrentMapShard{
			Items:         make(map[string][]byte),
			KeyExpiration: NewExpireQueue()}
	}
	return m, nil
}

// TODO??
func MakeStorageFromDump(dump []byte) (ConcurrentMap, error) {
	// deserialize into StorageDump
	m := make(ConcurrentMap, 0)

	// TODO
	return m, nil
}

func (c *ConcurrentMap) Get(key string) ([]byte, bool) {
	shard, ok := c.getShard(key)
	if !ok {
		return nil, false
	}

	shard.Lock()
	value, ok := shard.Items[key]
	shard.Unlock()
	return value, ok
}

func (c *ConcurrentMap) Set(key string, value []byte, ttl time.Duration) bool {
	shard, ok := c.getShard(key)
	if !ok || !validValue(value) {
		return false
	}
	shard.Lock()
	shard.Items[key] = value
	shard.KeyExpiration.InsertKey(key, ttl)
	shard.Unlock()
	return true
}

func (c *ConcurrentMap) Remove(key string) {
	shard, ok := c.getShard(key)
	if !ok {
		return
	}
	shard.Lock()
	delete(shard.Items, key)
	shard.Unlock()
}

func (c ConcurrentMap) Keys() []string {
	numShards := len(c)
	resChan := make(chan []string, numShards)

	// run key collectors
	for i := 0; i < numShards; i++ {
		go func(shardIndex int) {
			shard := c[shardIndex]
			shard.Lock()

			//fmt.Printf("goroutine #%d, lock shard\n", shardIndex)

			shardKeys := shard.getShardKeys()

			//fmt.Printf("goroutine #%d, get key chunk: %v\n", i, shardKeys)

			shard.Unlock()
			resChan <- shardKeys
		}(i)
	}

	// gather results

	//fmt.Println("waiting for Keys result...")

	keys := make([]string, 0)
	for i := 0; i < numShards; i++ {
		keyChunk := <-resChan

		//fmt.Printf("get key chunk #%d\n", i)

		keys = append(keys, keyChunk...)
	}

	//fmt.Println("ok, all keys are gathered!")

	return keys
}

/* internals */

// Return shard for given key
func (c ConcurrentMap) getShard(key string) (*ConcurrentMapShard, bool) {
	if !validKey(key) {
		return nil, false
	}
	return c[uint(utils.FNVSum64(key))%uint(len(c))], true
}

func validKey(key string) bool {
	if len(key) > KEY_MAX_LEN {
		return false
	}
	return true
}

func validValue(value []byte) bool {
	if len(value) > VALUE_MAX_SIZE {
		return false
	}
	return true
}

func (c *ConcurrentMapShard) getShardKeys() []string {
	i := 0
	//c.Lock()
	keys := make([]string, len(c.Items))
	for k, _ := range c.Items {
		keys[i] = k
		i++
	}
	//c.Unlock()
	return keys
}

////////////////////////////////

/*
func BootstrapStorage(dumpFile string, master bool, numShards int) (ConcurrentMap, error) {
	//
	if master {
		// TODO defer start slave-server

		// TODO read file
		snapshot := []byte{}

		if len(snapshot) > 0 {
			dump, err := decodeSnapshot(snapshot)
			if err == nil {
				return MakeStorageFromDump(dump)
			}
			// TODO write decode failure to log
		}
		// cannot restore data
		return MakeStorageEmpty(numShards)
	} else {
		// slave

	}
	return nil
}

func readSnapshotFromFile(filename string) ([]byte, error) {
	// TODO
	return nil, nil
}

func readSnapshotFromMaster(addr string) ([]byte, error) {
	// TODO
	return nil, nil
}

func decodeSnapshot(snapshot []byte) (StorageDump, error) {
	// TODO deserialize

	return StorageDump{}, nil
}
*/
