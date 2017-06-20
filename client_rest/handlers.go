package client_rest

import (
	"encoding/json"
	"net/http"
	"time"
)

const (
	KEY_TTL_MIN = 5
	KEY_TTL_MAX = 31 * 7 * 24 * 3600
)

/* request handlers */

func ResourceNotFound(w http.ResponseWriter, r *http.Request) {

	// TODO remove
	logger.Debugf("bad request: %s %s", r.Method, r.RequestURI)

	var response = ErrorResponse{
		Code:   ERR_CODE_ENDPOINT_NOT_FOUND,
		Reason: "endpoint not found"}

	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(response)
}

func GetItemHandler(w http.ResponseWriter, r *http.Request) {
	req, ok := readItemRequest(r.Body)
	if !ok {
		sendErrorResponse(w, http.StatusBadRequest, ERR_CODE_BAD_REQ, "cannot decode request")
		return
	}

	// TODO remove
	logger.Debugf("get item: %+v", req)

	// validate
	if req.Key == "" {
		sendErrorResponse(w, http.StatusBadRequest, ERR_CODE_NO_KEY_PROVIDED, "no key provided")
		return
	}

	store := GetStorageFromContext(r.Context())
	if req.SubKey != "" {
		// get item from value dictionary
		value, ok := GetDictItem(store, req.Key, req.SubKey)
		if ok {
			sendItemResponse(w, http.StatusOK, &CacheItem{Key: req.Key, SubKey: req.SubKey, Value: value})
		} else {
			sendErrorResponse(w, http.StatusBadRequest, ERR_CODE_NO_VALUE_FOUND, "value not found")
		}
	} else if req.SubIndex != 0 {
		// get item from value list
		// NOTE: element indexing in request starts from 1!
		value, ok := GetListItem(store, req.Key, req.SubIndex-1)
		if ok {
			sendItemResponse(w, http.StatusOK, &CacheItem{Key: req.Key, SubIndex: req.SubIndex, Value: value})
		} else {
			sendErrorResponse(w, http.StatusBadRequest, ERR_CODE_NO_VALUE_FOUND, "value not found")
		}
	} else {
		// get entire value
		value, ok := store.Get(req.Key)
		if ok {
			sendItemResponse(w, http.StatusOK, &CacheItem{Key: req.Key, Value: value})
		} else {
			sendErrorResponse(w, http.StatusBadRequest, ERR_CODE_NO_VALUE_FOUND, "value not found")
		}
	}
}

func SetItemHandler(w http.ResponseWriter, r *http.Request) {
	req, ok := readItemRequest(r.Body)
	if !ok {
		sendErrorResponse(w, http.StatusBadRequest, ERR_CODE_BAD_REQ, "cannot decode request")
		return
	}

	// TODO remove
	logger.Debugf("set item: %+v", req)

	// validate
	if req.Key == "" {
		sendErrorResponse(w, http.StatusBadRequest, ERR_CODE_NO_KEY_PROVIDED, "no key provided")
		return
	} else if len(req.Value) < 1 {
		sendErrorResponse(w, http.StatusBadRequest, ERR_CODE_NO_VALUE_PROVIDED, "no value provided")
		return
	} else if req.TTL < KEY_TTL_MIN || req.TTL > KEY_TTL_MAX {
		sendErrorResponse(w, http.StatusBadRequest, ERR_CODE_BAD_KEY_TTL, "bad key TTL")
		return
	}

	store := GetStorageFromContext(r.Context())
	if store.Set(req.Key, req.Value, time.Duration(req.TTL)*time.Second) {
		sendItemResponse(w, http.StatusCreated, req)
	} else {
		sendErrorResponse(w, http.StatusBadRequest, ERR_CODE_CANNOT_SET_KEY, "cannot save provided data")
	}
}

func RemoveItemHandler(w http.ResponseWriter, r *http.Request) {
	req, ok := readItemRequest(r.Body)
	if !ok {
		sendErrorResponse(w, http.StatusBadRequest, ERR_CODE_BAD_REQ, "cannot decode request")
		return
	}

	// TODO remove
	logger.Debugf("remove item: %+v", req)

	// validate
	if req.Key == "" {
		sendErrorResponse(w, http.StatusBadRequest, ERR_CODE_NO_KEY_PROVIDED, "no key provided")
		return
	}

	store := GetStorageFromContext(r.Context())
	store.Remove(req.Key)
	w.WriteHeader(http.StatusNoContent)
}

func GetKeysHandler(w http.ResponseWriter, r *http.Request) {
	req, ok := readKeysRequest(r.Body)
	store := GetStorageFromContext(r.Context())
	if !ok || req.Mask == "" {
		// get all keys
		sendKeysResponse(w, http.StatusOK, &KeysModel{Mask: "*", Keys: store.Keys()})
	} else {
		// get keys by mask
		keys, ok := store.KeysMask(req.Mask)
		if ok {
			sendKeysResponse(w, http.StatusOK, &KeysModel{Mask: req.Mask, Keys: keys})
		} else {
			sendErrorResponse(w, http.StatusBadRequest, ERR_CODE_BAD_KEY_MASK, "bad key mask")
		}
	}
}

/* helpers */

func sendErrorResponse(w http.ResponseWriter, header_status int, err_code int, reason string) {
	var response = ErrorResponse{
		Code:   err_code,
		Reason: reason}

	w.WriteHeader(header_status)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Errorf("cannot encode error message, code: %d, message: %s", err_code, reason)
	}
}

func sendItemResponse(w http.ResponseWriter, header_status int, itemResponse *CacheItem) {
	w.WriteHeader(header_status)
	if !writeItemResponse(w, itemResponse) {
		logger.Errorf("cannot encode item response: %+v", itemResponse)
	}
}

func sendKeysResponse(w http.ResponseWriter, header_status int, keysResponse *KeysModel) {
	w.WriteHeader(header_status)
	if !writeKeysResponse(w, keysResponse) {
		logger.Errorf("cannot encode item response: %+v", keysResponse)
	}
}
