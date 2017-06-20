package client_rest

import (
	"encoding/json"
	"github.com/dgtony/gcache/storage"
	"io"
)

/* User data serialization/deserialization */

func readItemRequest(r io.Reader) (*CacheItem, bool) {
	var req CacheItem
	if err := json.NewDecoder(r).Decode(&req); err != nil {
		return nil, false
	}
	return &req, true
}

func writeItemResponse(w io.Writer, item *CacheItem) bool {
	if err := json.NewEncoder(w).Encode(item); err != nil {
		return false
	}
	return true
}

func readKeysRequest(r io.Reader) (*KeysModel, bool) {
	var req KeysModel
	if err := json.NewDecoder(r).Decode(&req); err != nil {
		return nil, false
	}
	return &req, true
}

func writeKeysResponse(w io.Writer, keys *KeysModel) bool {
	if err := json.NewEncoder(w).Encode(keys); err != nil {
		return false
	}
	return true
}

/* additional methods */

// get value list item with index
func GetListItem(s *storage.ConcurrentMap, key string, subIndex int) ([]byte, bool) {
	if subIndex < 0 {
		return nil, false
	}

	res, ok := s.Get(key)
	if !ok {
		return nil, false
	}

	//try to decode value into list
	var valueList []interface{}
	if err := json.Unmarshal(res, &valueList); err != nil {
		return nil, false
	}

	// check boundaries
	if subIndex >= len(valueList) {
		return nil, false
	}

	// encode picked value back to bytes
	encodedItem, err := json.Marshal(valueList[subIndex])
	if err != nil {
		return nil, false
	}

	return encodedItem, true
}

// get value dictionary item with key
func GetDictItem(s *storage.ConcurrentMap, key, subKey string) ([]byte, bool) {
	res, ok := s.Get(key)
	if !ok {
		return nil, false
	}

	// try to decode value into dictionary
	var valueDict map[string]interface{}

	if err := json.Unmarshal(res, &valueDict); err != nil {
		return nil, false
	}

	value, ok := valueDict[subKey]
	if !ok {
		return nil, false
	}

	encodedItem, err := json.Marshal(value)
	if err != nil {
		return nil, false
	}

	return encodedItem, true
}
