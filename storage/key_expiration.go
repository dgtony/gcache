package storage

import (
	"container/heap"
	"time"
)

/*
Note: ExpireQueue structure is not thread-safe and must be used only in core storage!
Use separate heap for each shard and change it with shard lock.
*/

type StorageKey struct {
	Key    string
	Expire int64
}

type ExpireQueue []*StorageKey

func (q ExpireQueue) Len() int {
	return len(q)
}

func (q ExpireQueue) Less(i, j int) bool {
	// lowest expiration time first
	return q[i].Expire < q[j].Expire
}

func (q ExpireQueue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}

func (q *ExpireQueue) Push(k interface{}) {
	*q = append(*q, k.(*StorageKey))
}

func (q *ExpireQueue) Pop() interface{} {
	old := *q
	n := len(old)
	k := old[n-1]
	*q = old[0 : n-1]
	return k
}

/* queue methods */

func NewExpireQueue() ExpireQueue {
	return make([]*StorageKey, 0)
}

func (q *ExpireQueue) InsertKey(key string, ttl time.Duration) {
	keyExpiration := time.Now().Add(ttl).UnixNano()
	item := &StorageKey{Key: key, Expire: keyExpiration}
	heap.Push(q, item)
}

// return: (someKeysAreExpiredFlag, expiredKeys)
func (q *ExpireQueue) GetExpiredKeys() (bool, []string) {
	if q.Len() < 1 {
		return false, nil
	}

	var item *StorageKey
	expiredKeys := make([]string, 0)
	checkTime := time.Now().UnixNano()
	for {
		// get element
		item = heap.Pop(q).(*StorageKey)
		if item.Expire < checkTime {
			expiredKeys = append(expiredKeys, item.Key)
		} else {
			// insert active key back
			heap.Push(q, item)
			break
		}
	}

	// do not return empty lists
	if len(expiredKeys) > 0 {
		return true, expiredKeys
	}
	return false, nil
}
