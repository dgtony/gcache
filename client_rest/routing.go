package client_rest

import (
	"context"
	"github.com/dgtony/gcache/storage"
	"github.com/dgtony/gcache/utils"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
)

const (
	CTX_STORAGE_KEY = 1
)

type Route struct {
	Name     string
	Method   string
	Pattern  string
	HandlerF http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{
		Name:     "GetItem",
		Method:   "GET",
		Pattern:  "item",
		HandlerF: GetItemHandler},

	Route{
		Name:     "SetItem",
		Method:   "POST",
		Pattern:  "item",
		HandlerF: SetItemHandler},

	Route{
		Name:     "RemoveItem",
		Method:   "DELETE",
		Pattern:  "item",
		HandlerF: RemoveItemHandler},

	Route{
		Name:     "GetKeys",
		Method:   "GET",
		Pattern:  "keys",
		HandlerF: GetKeysHandler}}

func supplementRoute(route string, conf *utils.Config) string {
	var elems []string
	trimmedRoute := strings.TrimSpace(route)
	if len(trimmedRoute) == 0 || trimmedRoute == "/" {
		elems = []string{"", conf.ClientHTTP.RoutePrefix}
	} else {
		elems = []string{"", conf.ClientHTTP.RoutePrefix, trimmedRoute}
	}
	return strings.Join(elems, "/")
}

func wrapContextEnv(next http.HandlerFunc, store *storage.ConcurrentMap) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, CTX_STORAGE_KEY, store)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetStorageFromContext(ctx context.Context) *storage.ConcurrentMap {
	return ctx.Value(CTX_STORAGE_KEY).(*storage.ConcurrentMap)
}

func NewRouter(conf *utils.Config, store *storage.ConcurrentMap) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		// disable data changing endpoints on slave nodes
		if conf.Replication.NodeRole == "slave" && (route.Name == "SetItem" || route.Name == "RemoveItem") {
			continue
		}

		var handler http.HandlerFunc = route.HandlerF
		wrapped := wrapContextEnv(handler, store)
		fullRoute := supplementRoute(route.Pattern, conf)

		router.
			Methods(route.Method).
			Path(fullRoute).
			Name(route.Name).
			Handler(wrapped)
	}

	router.NotFoundHandler = http.HandlerFunc(ResourceNotFound)
	return router
}
