package agent

import (
	"fmt"
	"net/http"

	"github.com/doublemo/baa/cores/crypto/token"
	"github.com/doublemo/baa/kits/agent/router"
	"github.com/gorilla/mux"
)

// HttpRouterV1Config http v1
type HttpRouterV1Config struct {

	// CSRFSecret csrf key
	CSRFSecret string `alias:"csrf" default:"7581BDD8E8DA3839"`

	// CommandSecret 命令解密key
	CommandSecret string `alias:"commandSecret" default:"7581BDD8E8DA3839"`

	// MaxQureyLength 最大http query 长度
	MaxQureyLength int `alias:"maxQureyLength" default:"1024"`

	// MaxBytesReader 最大body
	MaxBytesReader int64 `alias:"maxBytesReader" default:"33554432"`
}

func httpRouterV1(r *mux.Router, c RouterConfig) {
	v1 := r.PathPrefix("/v1").Subrouter()
	v1.Use(csrfMethodMiddleware(c.HttpConfigV1))

	v1s := v1.PathPrefix("/x").Subrouter()
	v1s.Use(authenticationMiddleware)
	v1s.Handle("/snid/{command}", router.NewCall(c.ServiceSnid, Logger(), router.CommandSecretCallOptions(c.HttpConfigV1.CommandSecret))).Methods("POST")
}

func authenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Session-Token")
		if token == "" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
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
			fmt.Println("X-CSRF-Token", tks.Marshal(c.CSRFSecret), tks, token)
			if token != tks.Marshal(c.CSRFSecret) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, req)
		})
	}
}
