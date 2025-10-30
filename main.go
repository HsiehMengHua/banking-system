package main

import (
	"banking-system/router"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	router.Setup(r)
	r.Run() // listens on 0.0.0.0:8080 by default
}
