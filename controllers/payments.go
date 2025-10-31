package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type PaymentController interface {
	Deposit(c *gin.Context)
}

type paymentController struct {
}

func NewPaymentController() PaymentController {
	return &paymentController{}
}

// @Summary      Initiate a Deposit Transaction
// @Description  Creates a new PENDING transaction and redirects the user to the Payment Service Provider (PSP) for payment completion.
// @Tags         payments
// @Accept       json
// @Produce      json
// @Router       /payments/deposit [post]
func (*paymentController) Deposit(c *gin.Context) {
	c.JSON(http.StatusOK, "deposit response")
}
