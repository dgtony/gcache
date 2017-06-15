package core

import (
	"container/heap"
	"testing"
	"time"
	//remove
	//"fmt"
)

func TestKeyExpireHeapStructure(t *testing.T) {
	keyExpireQueue := NewExpireQueue()

	// test keys
	keyExpireQueue.InsertKey("key1", 20*time.Second)
	keyExpireQueue.InsertKey("key2", 30*time.Second)
	keyExpireQueue.InsertKey("key3", 10*time.Second)

	// pull ordered
	orderedKeys := make([]*StorageKey, 3)
	for i := 0; i < 3; i++ {
		orderedKeys[i] = heap.Pop(&keyExpireQueue).(*StorageKey)
	}

	if orderedKeys[0].Key != "key3" || orderedKeys[1].Key != "key1" || orderedKeys[2].Key != "key2" {
		t.Error("wrong expiration time ordering")
	}
}

func TestKeyExpireMethods(t *testing.T) {
	keyExpireQueue := NewExpireQueue()

	// test keys
	keyExpireQueue.InsertKey("key1", 1*time.Millisecond)
	keyExpireQueue.InsertKey("key2", 2*time.Second)
	keyExpireQueue.InsertKey("key3", 3*time.Millisecond)

	// wait less than expiration time
	time.Sleep(100 * time.Microsecond)
	if ready, _ := keyExpireQueue.GetExpiredKeys(); ready {
		t.Error("keys reported expired before real expiration")
	}

	// wait a bit more
	time.Sleep(10 * time.Millisecond)
	ready, expiredKeys := keyExpireQueue.GetExpiredKeys()

	if !ready {
		t.Error("expired keys are not reported")
	}
	if len(expiredKeys) != 2 || expiredKeys[0] != "key1" || expiredKeys[1] != "key3" {
		t.Error("wrong list of expired keys")
	}

}

func TestKeyExpireMassInsert(t *testing.T) {
	keyExpireQueue := NewExpireQueue()

	// insert short-living keys
	keysToInsert := 100
	for i := 0; i < keysToInsert; i++ {
		keyExpireQueue.InsertKey(string(i), 1*time.Microsecond)
	}
	// and one more
	keyExpireQueue.InsertKey("longlive", 1*time.Second)

	time.Sleep(10 * time.Microsecond)
	ready, expired := keyExpireQueue.GetExpiredKeys()
	if !ready || len(expired) != keysToInsert {
		t.Error("get expired keys failed: no short-living expired reported")
	}

	// long-living keys
	for i := 0; i < keysToInsert; i++ {
		keyExpireQueue.InsertKey(string(i), 1*time.Second)
	}
	if ready, _ := keyExpireQueue.GetExpiredKeys(); ready {
		t.Error("get expired keys failed: some of long-living expired")
	}
}
