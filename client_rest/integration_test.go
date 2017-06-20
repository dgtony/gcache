package client_rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dgtony/gcache/replicator"
	"github.com/dgtony/gcache/utils"
	"io/ioutil"
	"math/rand"
	"net/http"
	"testing"
)

const (
	symbols = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var decodedErr ErrorResponse
var decodedItem CacheItem
var decodedKeys KeysModel
var jsonPayload []byte

func TestClientRESTAPIBasic(t *testing.T) {
	routePrefix := "test"
	conf := getTestConfig(2, routePrefix)
	srv := startTestServer(conf)
	defer srv.Shutdown(nil)

	// get key with empty payload
	checkRespError(t, conf, "GET", "item", nil, http.StatusBadRequest, ERR_CODE_BAD_REQ)

	// get non-existing key
	jsonPayload = []byte(`{"key":"non_existent"}`)
	checkRespError(t, conf, "GET", "item", jsonPayload, http.StatusBadRequest, ERR_CODE_NO_VALUE_FOUND)

	// set new string key
	jsonPayload = []byte(`{"key":"testkey", "value": "testval", "ttl": 60}`)
	keyExp := "testkey"
	valExp := []byte("\"testval\"")
	item := checkRespItem(t, conf, "POST", "item", jsonPayload, http.StatusCreated)
	if item.Key != keyExp || !utils.CompareByteSlices(item.Value, valExp) {
		t.Errorf("expected => k: %s, v: %v | get => k: %s, v: %v", keyExp, valExp, item.Key, item.Value)
	}

	// remove key
	jsonPayload = []byte(`{"key":"testkey"}`)
	code, body, err := makeRequest(conf, "DELETE", "item", jsonPayload)
	if err != nil {
		t.Errorf("make request: %s", err)
	}
	if code != http.StatusNoContent || len(body) > 0 {
		t.Errorf("unexpected response => status: %d, response: %v", code, body)
	}

	// get key after removal
	jsonPayload = []byte(`{"key":"testkey"}`)
	checkRespError(t, conf, "GET", "item", jsonPayload, http.StatusBadRequest, ERR_CODE_NO_VALUE_FOUND)

}

func TestClientRESTAPISetItem(t *testing.T) {
	routePrefix := "test"
	conf := getTestConfig(2, routePrefix)
	srv := startTestServer(conf)
	defer srv.Shutdown(nil)

	// no key
	jsonPayload = []byte(`{"value": "testval", "ttl": 3600}`)
	checkRespError(t, conf, "POST", "item", jsonPayload, http.StatusBadRequest, ERR_CODE_NO_KEY_PROVIDED)

	// no value
	jsonPayload = []byte(`{"key":"testkey", "ttl": 3600}`)
	checkRespError(t, conf, "POST", "item", jsonPayload, http.StatusBadRequest, ERR_CODE_NO_VALUE_PROVIDED)

	// no ttl
	jsonPayload = []byte(`{"key":"testkey", "value": "testval"}`)
	checkRespError(t, conf, "POST", "item", jsonPayload, http.StatusBadRequest, ERR_CODE_BAD_KEY_TTL)

	// too small ttl
	jsonPayload = []byte(`{"key":"testkey", "value": "testval", "ttl": 2}`)
	checkRespError(t, conf, "POST", "item", jsonPayload, http.StatusBadRequest, ERR_CODE_BAD_KEY_TTL)

	// enormously huge ttl
	jsonPayload = []byte(`{"key":"testkey", "value": "testval", "ttl": 20000000}`)
	checkRespError(t, conf, "POST", "item", jsonPayload, http.StatusBadRequest, ERR_CODE_BAD_KEY_TTL)
}

func TestClientRESTAPIKeys(t *testing.T) {
	routePrefix := "test"
	conf := getTestConfig(8, routePrefix)
	srv := startTestServer(conf)
	defer srv.Shutdown(nil)

	// generate random keys
	numTestKeys := 1
	originalKeys := generateKeys(numTestKeys, 512)
	value := []byte("\"someunusedvalue\"")

	// set keys
	for _, k := range originalKeys {
		jsonPayload, err := json.Marshal(CacheItem{Key: k, Value: value, TTL: 60})
		if err != nil {
			t.Errorf("key encoding: %s", err)
		}
		checkRespItem(t, conf, "POST", "item", jsonPayload, http.StatusCreated)
	}

	// get all keys
	receivedKeys := checkRespKeys(t, conf, "GET", "keys", nil, http.StatusOK)
	if len(receivedKeys.Keys) != numTestKeys {
		t.Errorf("key number doesn't match => expected: %d, get: %d", numTestKeys, len(receivedKeys.Keys))
	}

	// check keys
	if !utils.CompareStringSlicesUnordered(originalKeys, receivedKeys.Keys) {
		t.Error("original and received keys doesn't match")
	}

}

func TestClientRESTAPISubElements(t *testing.T) {
	routePrefix := "test"
	conf := getTestConfig(2, routePrefix)
	srv := startTestServer(conf)
	defer srv.Shutdown(nil)

	// save dict
	jsonPayload = []byte(`{"key":"testdict", "value": {"a": 1, "b": 2}, "ttl": 60}`)
	checkRespItem(t, conf, "POST", "item", jsonPayload, http.StatusCreated)

	// get dict subkey
	jsonPayload = []byte(`{"key":"testdict", "subkey": "b"}`)
	subDictItem := checkRespItem(t, conf, "GET", "item", jsonPayload, http.StatusOK)
	if !utils.CompareByteSlices(subDictItem.Value, []byte("2")) {
		t.Errorf("subdict get failed => received item: %+v", subDictItem)
	}

	// save list
	jsonPayload = []byte(`{"key":"testlist", "value": ["some", "words"], "ttl": 60}`)
	checkRespItem(t, conf, "POST", "item", jsonPayload, http.StatusCreated)

	// get list subindex
	jsonPayload = []byte(`{"key":"testlist", "subindex": 2}`)
	subListItem := checkRespItem(t, conf, "GET", "item", jsonPayload, http.StatusOK)
	if !utils.CompareByteSlices(subListItem.Value, []byte("\"words\"")) {
		t.Errorf("sublist get failed => received item: %+v", subListItem)
	}
}

/* helpers */

func checkRespError(t *testing.T, conf *utils.Config, method, endpoint string, jsonPayload []byte, expHTTPCode, expErrCode int) {
	code, body, err := makeRequest(conf, method, endpoint, jsonPayload)
	if err != nil {
		t.Errorf("make request: %s", err)
	}
	if err := json.Unmarshal(body, &decodedErr); err != nil {
		t.Errorf("decoding response: %s", err)
	}
	if code != expHTTPCode || decodedErr.Code != expErrCode {
		t.Errorf("unexpected response => status code: %d, error code: %d", code, decodedErr.Code)
	}
}

func checkRespItem(t *testing.T, conf *utils.Config, method, endpoint string, jsonPayload []byte, expHTTPCode int) CacheItem {
	code, body, err := makeRequest(conf, method, endpoint, jsonPayload)
	if err != nil {
		t.Errorf("make request: %s", err)
	}
	if err := json.Unmarshal(body, &decodedItem); err != nil {
		t.Errorf("decoding response: %s", err)
	}
	if code != expHTTPCode {
		t.Errorf("unexpected response status => code: %d", code)
	}
	return decodedItem
}

func checkRespKeys(t *testing.T, conf *utils.Config, method, endpoint string, jsonPayload []byte, expHTTPCode int) KeysModel {
	code, body, err := makeRequest(conf, method, endpoint, jsonPayload)
	if err != nil {
		t.Errorf("make request: %s", err)
	}
	if err := json.Unmarshal(body, &decodedKeys); err != nil {
		t.Errorf("decoding response: %s", err)
	}
	if code != expHTTPCode {
		t.Errorf("unexpected response status => code: %d", code)
	}
	return decodedKeys
}

func getTestConfig(numShards int, routePrefix string) *utils.Config {
	return &utils.Config{
		General: utils.GeneralSettings{
			LogLevel:  "debug",
			LogFormat: "short",
			LogOut:    "stdout"},
		Replication: utils.ReplicationSettings{
			NodeRole:             "standalone",
			RestoreCacheFromFile: false,
			SaveCacheToFile:      false,
			DumpUpdatePeriod:     10,
		},
		Storage: utils.StorageSettings{
			NumShards:               numShards,
			ExpiredKeyCheckInterval: 10},
		ClientHTTP: utils.ClientHTTPSettings{
			Port:        "12345",
			RoutePrefix: routePrefix,
			IdleTimeout: 60}}
}

func startTestServer(conf *utils.Config) *http.Server {
	utils.SetupLoggers(conf)
	_, store := replicator.RunReplicator(conf)
	return StartClientREST(conf, store)
}

// return status code, raw body and error
func makeRequest(conf *utils.Config, method, endpoint string, jsonPayload []byte) (int, []byte, error) {
	var req *http.Request
	url := buildURL(conf, endpoint)
	if jsonPayload != nil {
		req, _ = http.NewRequest(method, url, bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, _ = http.NewRequest(method, url, nil)
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return resp.StatusCode, body, nil
}

func buildURL(conf *utils.Config, endpoint string) string {
	return fmt.Sprintf("http://localhost:%s/%s/%s", conf.ClientHTTP.Port, conf.ClientHTTP.RoutePrefix, endpoint)
}

func randString(maxSize int) string {
	sliceSize := rand.Intn(maxSize)
	b := make([]byte, sliceSize)
	for i := range b {
		b[i] = symbols[rand.Int63()%int64(len(symbols))]
	}
	return string(b)
}

func generateKeys(numKeys, keyMaxSize int) []string {
	res := make([]string, numKeys)
	for i := 0; i < numKeys; i++ {
		res[i] = randString(keyMaxSize)
	}
	return res
}
