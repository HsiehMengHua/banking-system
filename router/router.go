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

	userRepo := repos.NewUserRepo()
	paymentCtrl := controllers.NewPaymentController(services.NewPaymentService(userRepo, repos.NewTransactionRepo(), psp.NewPaymentServiceProvider()))
	userCtrl := controllers.NewUserController(services.NewUserService(userRepo))

	api := r.Group("/api/v1")
	{
		api.POST("/user", userCtrl.Register)

		payments := api.Group("/payments")
		payments.POST("/deposit", paymentCtrl.Deposit)
		payments.POST("/withdraw", paymentCtrl.Withdraw)
		payments.POST("/transfer", paymentCtrl.Transfer)
		payments.POST("/confirm", middleware.VerifyPSPApiKey(), paymentCtrl.Confirm)
		payments.POST("/cancel", middleware.VerifyPSPApiKey(), paymentCtrl.Cancel)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	return r
}
