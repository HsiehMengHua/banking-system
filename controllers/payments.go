package controllers

import (
	"banking-system/models"
	"banking-system/psp"
	"banking-system/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PaymentController interface {
	Deposit(c *gin.Context)
	Confirm(c *gin.Context)
	Cancel(c *gin.Context)
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
// @Response     302  {object}  object  "Redirect to PSP payment page"
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

// @Summary      Confirm a Deposit Transaction
// @Description  Handles the confirmation callback from the Payment Service Provider (PSP) after a successful deposit.
// @Tags         payments
// @Accept       json
// @Response     200  {string}  string	"Deposit confirmed successfully"
// @Router       /payments/confirm [post]
func (ctrl *paymentController) Confirm(c *gin.Context) {
	var req psp.ConfirmRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request body or missing field: " + err.Error(),
		})
		return
	}

	ctrl.paymentSrv.Confirm(&req)
	c.Status(http.StatusOK)
}
func (ctrl *paymentController) Cancel(c *gin.Context) {}
