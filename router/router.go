package router

import (
	"github.com/gin-gonic/gin"
)

func Setup(r *gin.Engine) {
	r.StaticFile("/version", "./version.txt")

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
}
