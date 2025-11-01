package controllers

import (
	"banking-system/models"
	"banking-system/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PaymentController interface {
	Deposit(c *gin.Context)
}

type paymentController struct {
	paymentSrv services.PaymentService
}

func NewPaymentController(paymentSrv services.PaymentService) PaymentController {
	return &paymentController{
		paymentSrv: paymentSrv,
	}
}

// @Summary      Initiate a Deposit Transaction
// @Description  Creates a new PENDING transaction and redirects the user to the Payment Service Provider (PSP) for payment completion.
// @Tags         payments
// @Accept       json
// @Produce      json
// @Router       /payments/deposit [post]
func (ctrl *paymentController) Deposit(c *gin.Context) {
	var req models.DepositRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request body or missing field: " + err.Error(),
		})
		return
	}

	redirectUrl := ctrl.paymentSrv.Deposit(&req)
	c.Redirect(http.StatusFound, redirectUrl)
}
