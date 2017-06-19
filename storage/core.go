package storage

import (
	"errors"
	"github.com/dgtony/gcache/utils"
	"github.com/gobwas/glob"
	"github.com/op/go-logging"
	"sync"
	"time"
)

const (
	MAX_SHARDS = 4096
	// key length limit
	KEY_MAX_LEN = 2048
	// value size limit up to 10Mb
	VALUE_MAX_SIZE = 10485760
)

var logger *logging.Logger

func init_logger() {
	logger = utils.GetLogger("Storage")
}

type ConcurrentMap []*ConcurrentMapShard

type ConcurrentMapShard struct {
	Items         map[string][]byte
	KeyExpiration ExpireQueue
	sync.RWMutex
}

/* Storage methods */

func MakeStorageEmpty(conf *utils.Config) (*ConcurrentMap, error) {
	init_logger()

	numShards := conf.Storage.NumShards

	logger.Debugf("create storage with %d shards", numShards)

	if numShards < 1 || numShards > MAX_SHARDS {
		return nil, errors.New("wrong number of shards")
	}

	m := make(ConcurrentMap, numShards)
	for i := 0; i < numShards; i++ {
		m[i] = &ConcurrentMapShard{
			Items:         make(map[string][]byte),
			KeyExpiration: NewExpireQueue()}
	}

	cleanPeriod := time.Duration(conf.Storage.ExpiredKeyCheckInterval) * time.Second
	m.runExpKeyCleaning(cleanPeriod)
	return &m, nil
}

func MakeStorageFromDump(conf *utils.Config, snapshot []byte) (*ConcurrentMap, error) {
	init_logger()

	// decode snapshot
	storageDump, err := deserializeDump(snapshot)
	if err != nil {
		return nil, err
	}

	// create storage
	numShards := len(storageDump)
	m := make(ConcurrentMap, numShards)
	for i := 0; i < numShards; i++ {
		m[i] = &ConcurrentMapShard{
			Items:         storageDump[i].Items,
			KeyExpiration: storageDump[i].KeyExpiration}
	}
	cleanPeriod := time.Duration(conf.Storage.ExpiredKeyCheckInterval) * time.Second
	m.runExpKeyCleaning(cleanPeriod)
	return &m, nil
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
			shardKeys := shard.getShardKeys()
			shard.Unlock()
			resChan <- shardKeys
		}(i)
	}

	// gather results
	keys := make([]string, 0)
	for i := 0; i < numShards; i++ {
		keyChunk := <-resChan
		keys = append(keys, keyChunk...)
	}

	return keys
}

// return keys according to given mask
// use glob pattern matching
func (c ConcurrentMap) KeysMask(mask string) ([]string, bool) {
	g, err := glob.Compile(mask)
	if err != nil {
		return nil, false
	}
	filtered := make([]string, 0)
	for _, s := range c.Keys() {
		if g.Match(s) {
			filtered = append(filtered, s)
		}
	}
	return filtered, true
}

/* internals */

func (c ConcurrentMap) runExpKeyCleaning(cleanPeriod time.Duration) {
	for _, shard := range c {
		// run separate cleaner process for each shard
		go func(shard *ConcurrentMapShard) {
			for {
				shard.Lock()
				ok, expiredKeys := shard.KeyExpiration.GetExpiredKeys()
				if ok {
					for _, k := range expiredKeys {
						delete(shard.Items, k)
					}
				}
				shard.Unlock()
				time.Sleep(cleanPeriod)
			}
		}(shard)
	}
}

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

// do not use outside - not thread-safe!
func (c *ConcurrentMapShard) getShardKeys() []string {
	i := 0
	keys := make([]string, len(c.Items))
	for k, _ := range c.Items {
		keys[i] = k
		i++
	}
	return keys
}

////////////////////////////////

// WTD?

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
