package storage

import (
	"github.com/dgtony/gcache/utils"
	"testing"
	"time"
)

func makeTestDump() StorageDump {
	expireTime := time.Now().Add(time.Second).UnixNano()
	return []ShardDump{
		ShardDump{
			Items: map[string][]byte{
				"key11": []byte("value11"),
				"key12": []byte("value12")},
			KeyExpiration: []*StorageKey{
				&StorageKey{Key: "key11", Expire: expireTime},
				&StorageKey{Key: "key12", Expire: expireTime}}},
		ShardDump{
			Items: map[string][]byte{
				"key21": []byte("value21"),
				"key22": []byte("value22"),
				"key23": []byte("value23")},
			KeyExpiration: []*StorageKey{
				&StorageKey{Key: "key21", Expire: expireTime},
				&StorageKey{Key: "key22", Expire: expireTime},
				&StorageKey{Key: "key23", Expire: expireTime}}}}
}

func TestCoreDumpSerializeDeserialize(t *testing.T) {
	dumpOriginal := makeTestDump()
	ser, err := serializeDump(dumpOriginal)
	if err != nil {
		t.Errorf("dump serialization: %s", err)
	}

	dumpRestored, err := deserializeDump(ser)
	if err != nil {
		t.Errorf("dump deserialization: %s", err)
	}

	if !compareStorageDumps(dumpOriginal, dumpRestored) {
		t.Error("dump serialization is not isomorphic")
	}
}

/* internals */

func compareStorageDumps(d1, d2 StorageDump) bool {
	// number of shards
	if len(d1) != len(d2) {
		return false
	}

	for i := 0; i < len(d1); i++ {
		// compare stored values
		if !utils.CompareStringByteMaps(d1[i].Items, d2[i].Items) {
			return false
		}

		// compare key expirations
		if !compareExpireQueue(d1[i].KeyExpiration, d2[i].KeyExpiration) {
			return false
		}
	}

	return true
}

func compareExpireQueue(q1, q2 ExpireQueue) bool {
	if len(q1) != len(q2) {
		return false
	}

	for i := 0; i < len(q1); i++ {
		if q1[i].Key != q2[i].Key || q1[i].Expire != q2[i].Expire {
			return false
		}
	}
	return true
}
