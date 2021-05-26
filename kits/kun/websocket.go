// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/baa>

package kun

import (
	"errors"
	"fmt"
	"net/http"

	log "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/kits/kun/mem"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func websocketRouter(o *mem.Parameters) func(*gin.Engine) {
	return func(r *gin.Engine) {
		if o.Websocket == nil {
			return
		}

		var webSocketUpgrader websocket.Upgrader
		{
			webSocketUpgrader = websocket.Upgrader{
				ReadBufferSize:  o.Websocket.ReadBufferSize,
				WriteBufferSize: o.Websocket.WriteBufferSize,
				CheckOrigin: func(r *http.Request) bool {
					return true
				},
			}
		}

		r.GET("/ws", func(ctx *gin.Context) {
			if !ctx.IsWebsocket() {
				ctx.AbortWithError(http.StatusNotFound, errors.New("404 Not found"))
				return
			}

			webscoketHandler(ctx.Writer, ctx.Request, webSocketUpgrader, o)
		})
	}
}

func webscoketHandler(w http.ResponseWriter, req *http.Request, upgrader websocket.Upgrader, o *mem.Parameters) {
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Error(logger).Log("error", err)
		return
	}

	exit := make(chan struct{})
	fmt.Println(conn, exit)
}
