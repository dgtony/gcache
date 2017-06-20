package storage

import (
	"bytes"
	"encoding/gob"
)

type StorageDump []ShardDump
type ShardDump struct {
	Items         map[string][]byte
	KeyExpiration ExpireQueue
}

// Get current storage snapshot
func (c ConcurrentMap) DumpStorage() ([]byte, error) {
	numShards := len(c)
	fullDump := make([]ShardDump, numShards)
	for i, shard := range c {
		shard.Lock()
		fullDump[i].Items = copyShardItems(shard)
		fullDump[i].KeyExpiration = copyShardKeyExp(shard)
		shard.Unlock()
	}

	serialized, err := serializeDump(fullDump)
	if err != nil {
		return nil, err
	}

	return serialized, nil
}

// Restore entire storage from snapshot
// All stored data will be completely replaced
func (c *ConcurrentMap) RestoreFromDump(snapshot []byte) error {
	storageDump, err := deserializeDump(snapshot)
	if err != nil {
		return err
	}

	if len(*c) != len(storageDump) {
		// dunno how to recover in this case
		panic("cannot restore from dump: shard number doesn't match")
	}

	// restore each shard separately
	for i, shardDump := range storageDump {
		go func(shardIndex int, shardDump ShardDump) {
			oldShard := (*c)[shardIndex]
			oldShard.Lock()
			oldShard.Items = shardDump.Items
			oldShard.KeyExpiration = shardDump.KeyExpiration
			oldShard.Unlock()
		}(i, shardDump)
	}

	return nil
}

func copyShardItems(shard *ConcurrentMapShard) map[string][]byte {
	newShardItems := make(map[string][]byte)
	for k, v := range shard.Items {
		newShardItems[k] = v
	}
	return newShardItems
}

func copyShardKeyExp(shard *ConcurrentMapShard) ExpireQueue {
	keyExpLen := len(shard.KeyExpiration)
	newKeyExpirations := make([]*StorageKey, keyExpLen)

	for i, item := range shard.KeyExpiration {
		newKeyExpirations[i] = &StorageKey{Key: item.Key, Expire: item.Expire}
	}

	return newKeyExpirations
}

func serializeDump(dump StorageDump) ([]byte, error) {
	var buff bytes.Buffer
	err := gob.NewEncoder(&buff).Encode(dump)
	return buff.Bytes(), err
}

func deserializeDump(snapshot []byte) (StorageDump, error) {
	var dump StorageDump
	var buff bytes.Buffer
	buff.Write(snapshot)
	err := gob.NewDecoder(&buff).Decode(&dump)
	return dump, err
}
