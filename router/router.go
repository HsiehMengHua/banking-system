package router

import (
	"banking-system/controllers"
	"banking-system/docs"
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
	paymentCtrl := controllers.NewPaymentController(services.NewPaymentService(userRepo, repos.NewTransactionRepo(), psp.NewPSPFactory()))
	userCtrl := controllers.NewUserController(services.NewUserService(userRepo))

	api := r.Group("/api/v1")
	{
		{
			userApi := api.Group("/user")
			userApi.POST("", userCtrl.Register)
			userApi.POST("/login", userCtrl.Login)
		}

		{
			paymentApi := api.Group("/payments")
			paymentApi.POST("/deposit", paymentCtrl.Deposit)
			paymentApi.POST("/withdraw", paymentCtrl.Withdraw)
			paymentApi.POST("/transfer", paymentCtrl.Transfer)
			paymentApi.POST("/confirm", paymentCtrl.Confirm)
			paymentApi.POST("/cancel", paymentCtrl.Cancel)
		}
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	return r
}
