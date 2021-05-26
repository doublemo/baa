// Copyright (c) 2019 The baa Authors <https://github.com/doublemo/baa>

package kun

import (
	"net/http"
	"strconv"

	log "github.com/doublemo/baa/cores/log/level"
	"github.com/doublemo/baa/kits/kun/mem"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func httpRouter(o *mem.Parameters) func(*gin.Engine) {
	return func(r *gin.Engine) {
		// 定义命令路由
		if o.Runmode == "dev" {
			r.GET("/metrics", gin.WrapH(promhttp.Handler()))
		}

		r.GET("/v1/:service/:method", httpHandler(o, 1))
		r.POST("/v1/:service/:method", httpHandler(o, 1))
	}
}

func httpHandler(o *mem.Parameters, v int) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		service, err := strconv.ParseInt(ctx.Param("service"), 10, 32)
		if err != nil {
			ctx.JSON(http.StatusOK, "ddd")
			return
		}

		method, err := strconv.ParseInt(ctx.Param("method"), 10, 32)
		if err != nil {
			ctx.JSON(http.StatusOK, "")
			return
		}

		log.Debug(logger).Log("service", service, "method", method, "v", v)
		ctx.JSON(http.StatusOK, map[string]string{"code": "ttttt"})
	}
}
