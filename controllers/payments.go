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
	Withdraw(c *gin.Context)
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
// @Param        request body models.DepositRequest true "Deposit initiation details"
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

	redirectUrl, err := ctrl.paymentSrv.Deposit(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.Redirect(http.StatusFound, redirectUrl)
}

// @Summary      Initiate a Withdrawal Transaction
// @Description  Creates a new PENDING withdrawal transaction, deducts the amount from wallet balance, and sends request to PSP for processing.
// @Tags         payments
// @Accept       json
// @Param        request body models.WithdrawRequest true "Withdrawal initiation details"
// @Response     200  {object}  object  "Withdrawal initiated successfully"
// @Response     400  {object}  object  "Bad request - validation error or insufficient balance"
// @Router       /payments/withdraw [post]
func (ctrl *paymentController) Withdraw(c *gin.Context) {
	var req models.WithdrawRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request body or missing field: " + err.Error(),
		})
		return
	}

	if err := ctrl.paymentSrv.Withdraw(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.Status(http.StatusOK)
}

// @Summary      Confirm a Deposit Transaction
// @Description  Handles the confirmation callback from the Payment Service Provider (PSP) after a successful deposit.
// @Tags         payments
// @Accept       json
// @Param        X-PSP-API-Key header string true "PSP API Key"
// @Param        request body psp.PayInResponse true "Confirmation callback from PSP"
// @Response     200  {string}  string	"Deposit confirmed successfully"
// @Response     401  {object}  object	"Unauthorized - Invalid API key"
// @Router       /payments/confirm [post]
func (ctrl *paymentController) Confirm(c *gin.Context) {
	var req psp.ConfirmRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request body or missing field: " + err.Error(),
		})
		return
	}

	if err := ctrl.paymentSrv.Confirm(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.Status(http.StatusOK)
}

// @Summary      Cancel a Deposit Transaction
// @Description  Handles the cancellation callback from the Payment Service Provider (PSP) when a deposit is cancelled.
// @Tags         payments
// @Accept       json
// @Param        X-PSP-API-Key header string true "PSP API Key"
// @Param        request body psp.CancelRequest true "Cancellation callback from PSP"
// @Response     200  {string}  string	"Deposit cancelled successfully"
// @Response     401  {object}  object	"Unauthorized - Invalid API key"
// @Router       /payments/cancel [post]
func (ctrl *paymentController) Cancel(c *gin.Context) {
	var req psp.CancelRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request body or missing field: " + err.Error(),
		})
		return
	}

	if err := ctrl.paymentSrv.Cancel(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.Status(http.StatusOK)
}
