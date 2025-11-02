package router

import (
	"banking-system/controllers"
	"banking-system/docs"
	"banking-system/middleware"
	"banking-system/psp"
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

	ctrl := controllers.NewPaymentController(services.NewPaymentService(repos.NewUserRepo(), repos.NewTransactionRepo(), psp.NewPaymentServiceProvider()))

	api := r.Group("/api/v1")
	{
		payments := api.Group("/payments")
		payments.POST("/deposit", ctrl.Deposit)
		payments.POST("/withdraw", ctrl.Withdraw)
		payments.POST("/confirm", middleware.VerifyPSPApiKey(), ctrl.Confirm)
		payments.POST("/cancel", middleware.VerifyPSPApiKey(), ctrl.Cancel)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	return r
}
