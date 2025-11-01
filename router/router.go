package router

import (
	"banking-system/controllers"
	"banking-system/docs"
	"banking-system/repos"
	"banking-system/services"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Setup() *gin.Engine {
	r := gin.Default()
	docs.SwaggerInfo.BasePath = "/api/v1"
	r.StaticFile("/version", "./version.txt")

	ctrl := controllers.NewPaymentController(services.NewPaymentService(repos.NewUserRepo(), repos.NewTransactionRepo()))

	{
		payments := r.Group("/payments")
		payments.POST("/deposit", ctrl.Deposit)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	return r
}
