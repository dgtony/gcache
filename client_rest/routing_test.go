package client_rest

import (
	"fmt"
	"testing"
)

func TestClientRESTRoutingSupplementRoute(t *testing.T) {
	testPrefixes := []string{
		"",
		"test",
		"test/prefix",
	}
	for _, routePrefix := range testPrefixes {
		conf := getTestConfig(1, routePrefix)
		testRoutes := map[string]string{
			"":         fmt.Sprintf("/%s", routePrefix),
			" ":        fmt.Sprintf("/%s", routePrefix),
			"    ":     fmt.Sprintf("/%s", routePrefix),
			"\r\n":     fmt.Sprintf("/%s", routePrefix),
			" \r\n \n": fmt.Sprintf("/%s", routePrefix),
			"/":        fmt.Sprintf("/%s", routePrefix),
			"key":      fmt.Sprintf("/%s/key", routePrefix),
			" key":     fmt.Sprintf("/%s/key", routePrefix),
			" key ":    fmt.Sprintf("/%s/key", routePrefix),
			"key\n":    fmt.Sprintf("/%s/key", routePrefix),
			"key\r\n":  fmt.Sprintf("/%s/key", routePrefix),
			"key/op":   fmt.Sprintf("/%s/key/op", routePrefix),
		}

		for route, result := range testRoutes {
			if supplementRoute(route, conf) != result {
				t.Errorf("wrong route supplement, expected: %s, get: %s", result, supplementRoute(route, conf))
			}
		}
	}
}
