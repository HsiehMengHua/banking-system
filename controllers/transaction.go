package controllers

import (
	"banking-system/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type TransactionController interface {
	GetByUserID(c *gin.Context)
}

type transactionController struct {
	transactionSrv services.TransactionService
}

func NewTransactionController(transactionSrv services.TransactionService) TransactionController {
	return &transactionController{
		transactionSrv: transactionSrv,
	}
}

// @Summary      Get user transactions
// @Description  Retrieves all transactions for a user within the last 6 months
// @Tags         transactions
// @Accept       json
// @Param        user_id path int true "User ID"
// @Success      200  {array}  models.TransactionResponse  "List of transactions"
// @Response     400  {object}  object  "Bad request - invalid user ID"
// @Router       /transactions/user/{user_id} [get]
func (ctrl *transactionController) GetByUserID(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid user ID",
		})
		return
	}

	const months = 6
	cutoffDate := time.Now().AddDate(0, -months, 0)
	transactions, err := ctrl.transactionSrv.GetByUserID(uint(userID), cutoffDate)
	if err != nil {
		log.Errorf("Failed to get transactions for user %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to retrieve transactions",
		})
		return
	}

	c.JSON(http.StatusOK, transactions)
}
