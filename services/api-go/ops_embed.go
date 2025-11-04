// services/api-go/ops_embed.go
package main

import (
	_ "embed"

	"github.com/gin-gonic/gin"
)

//go:embed ops.html
var opsHTML []byte

func mountOps(r *gin.Engine) {
	r.GET("/ops", func(c *gin.Context) {
		c.Data(200, "text/html; charset=utf-8", opsHTML)
	})
}
