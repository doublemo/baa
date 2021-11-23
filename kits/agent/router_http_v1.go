package agent

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/doublemo/baa/cores/crypto/token"
	coressd "github.com/doublemo/baa/cores/sd"
	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/internal/sd"
	"github.com/doublemo/baa/kits/agent/router"
	"github.com/gorilla/mux"
	"google.golang.org/grpc/resolver"
)

type (
	// HttpRouterV1Config http v1
	HttpRouterV1Config struct {

		// CSRFSecret csrf key
		CSRFSecret string `alias:"csrf" default:"7581BDD8E8DA3839"`

		// CommandSecret 命令解密key
		CommandSecret string `alias:"commandSecret" default:"7581BDD8E8DA3839"`

		// MaxQureyLength 最大http query 长度
		MaxQureyLength int `alias:"maxQureyLength" default:"1024"`

		// MaxBytesReader 最大body
		MaxBytesReader int64 `alias:"maxBytesReader" default:"33554432"`

		// Routes 路由控制
		Routes []HttpRouter `alias:"routes"`
	}

	HttpRouter struct {
		Path          string         `alias:"path"`
		Authorization bool           `alias:"authorization"`
		Method        string         `alias:"method"`
		Config        conf.RPCClient `alias:"config"`
		Commands      []int32        `alias:"commands"`
	}
)

func httpRouterV1(r *mux.Router, c RouterConfig) {
	v1 := r.PathPrefix("/v1").Subrouter()
	v1.Use(csrfMethodMiddleware(c.HttpConfigV1))

	v1s := v1.PathPrefix("/x").Subrouter()
	v1s.Use(authenticationMiddleware)

	var subrouter *mux.Router
	for _, route := range c.HttpConfigV1.Routes {
		if m := resolver.Get(route.Config.Name); m == nil {
			resolver.Register(coressd.NewResolverBuilder(route.Config.Name, route.Config.Group, sd.Endpointer()))
		}

		path := route.Path
		if !strings.HasSuffix(path, "/") {
			path += "/"
		}

		subrouter = v1
		if route.Authorization {
			subrouter = v1s
		}

		opts := []router.CallOptions{
			router.CommandSecretCallOptions(c.HttpConfigV1.CommandSecret),
			router.AllowCommandsCallOptions(route.Commands...),
		}

		subrouter.Handle(route.Path+"{command}", router.NewCall(route.Config, Logger(), opts...)).Methods(route.Method)
	}
}

func authenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Session-Token")
		if token == "" {
			urlQuery := r.URL.Query()
			token = urlQuery.Get("token")
			if token == "" {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
		}

		info, err := authenticateToken(token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		values := r.URL.Query()
		values.Add("UserID", strconv.FormatUint(info.UserID, 10))
		values.Add("AccountID", strconv.FormatUint(info.ID, 10))
		r.URL.RawQuery = values.Encode()
		next.ServeHTTP(w, r)
	})
}

// X-CSRF-Token
func csrfMethodMiddleware(c HttpRouterV1Config) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			urlQuery := req.URL.Query()
			vars := mux.Vars(req)
			tks := token.NewTKS()
			if m, ok := vars["command"]; ok {
				tks.Push("command=" + m)
			}

			time := urlQuery.Get("t")
			tks.Push("t=" + time)
			token := req.Header.Get("X-CSRF-Token")
			if token == "" {
				token = urlQuery.Get("X-CSRF-Token")
				tks.Push("X-CSRF-Token=" + token)
			}
			fmt.Println("X-CSRF-Token", tks.Marshal(c.CSRFSecret), tks, token)
			if token != tks.Marshal(c.CSRFSecret) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, req)
		})
	}
}
