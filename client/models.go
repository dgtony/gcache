package client

import (
	"encoding/json"
)

const (
	// general errors
	ERR_CODE_ENDPOINT_NOT_FOUND = 1
	ERR_CODE_BAD_REQ            = 2

	// request format errors
	ERR_CODE_NO_KEY_PROVIDED   = 10
	ERR_CODE_NO_VALUE_PROVIDED = 11
	ERR_CODE_BAD_KEY_TTL       = 12
	ERR_CODE_BAD_KEY_MASK      = 13

	// response errors
	ERR_CODE_NO_VALUE_FOUND = 21
	ERR_CODE_CANNOT_SET_KEY = 22
)

type CacheItem struct {
	Key      string          `json:"key"`
	Value    json.RawMessage `json:"value,omitempty"`
	SubKey   string          `json:"subkey,omitempty"`
	SubIndex int             `json:"subindex,omitempty"`
	TTL      int             `json:"ttl,omitempty"`
}

type KeysModel struct {
	Mask string   `json:"mask"`
	Keys []string `json:"keys"`
}

type ErrorResponse struct {
	Code   int    `json:"code"`
	Reason string `json:"reason"`
}
