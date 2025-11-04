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
	bankAccountCtrl := controllers.NewBankAccountController(services.NewBankAccountService(repos.NewBankAccountRepo()))

	api := r.Group("/api/v1")
	{
		{
			userApi := api.Group("/user")
			userApi.POST("", userCtrl.Register)
			userApi.POST("/login", userCtrl.Login)
			userApi.GET("/:user_id", userCtrl.GetByID)
		}

		{
			paymentApi := api.Group("/payments")
			paymentApi.POST("/deposit", paymentCtrl.Deposit)
			paymentApi.POST("/withdraw", paymentCtrl.Withdraw)
			paymentApi.POST("/transfer", paymentCtrl.Transfer)
			paymentApi.POST("/confirm", paymentCtrl.Confirm)
			paymentApi.POST("/cancel", paymentCtrl.Cancel)
		}

		{
			bankAccountApi := api.Group("/bank-accounts")
			bankAccountApi.GET("/user/:userId", bankAccountCtrl.GetByUserID)
			bankAccountApi.POST("", bankAccountCtrl.Create)
			bankAccountApi.GET("", bankAccountCtrl.GetAll)
			bankAccountApi.GET("/:id", bankAccountCtrl.GetByID)
			bankAccountApi.PUT("/:id", bankAccountCtrl.Update)
			bankAccountApi.DELETE("/:id", bankAccountCtrl.Delete)
		}
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	return r
}
