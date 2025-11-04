package controllers

import (
	"banking-system/entities"
	"banking-system/models"
	"banking-system/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type BankAccountController interface {
	Create(c *gin.Context)
	GetByID(c *gin.Context)
	GetAll(c *gin.Context)
	GetByUserID(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

type bankAccountController struct {
	bankAccountSrv services.BankAccountService
}

func NewBankAccountController(bankAccountSrv services.BankAccountService) BankAccountController {
	return &bankAccountController{
		bankAccountSrv: bankAccountSrv,
	}
}

// @Summary      Create a new bank account
// @Description  Creates a new bank account for the authenticated user
// @Tags         bank-accounts
// @Accept       json
// @Param        X-User-ID header string true "User ID"
// @Param        request body models.CreateBankAccountRequest true "Bank account creation details"
// @Success      201  {object}  models.BankAccountResponse  "Bank account created successfully"
// @Response     400  {object}  object  "Bad request - validation error"
// @Router       /bank-account [post]
func (ctrl *bankAccountController) Create(c *gin.Context) {
	var req models.CreateBankAccountRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request body or missing field: " + err.Error(),
		})
		return
	}

	userID, err := getUserIDFromHeader(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
		return
	}

	req.UserID = userID

	response, err := ctrl.bankAccountSrv.Create(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// @Summary      Get a bank account by ID
// @Description  Retrieves a specific bank account by ID for the authenticated user
// @Tags         bank-accounts
// @Accept       json
// @Param        X-User-ID header string true "User ID"
// @Param        id path int true "Bank Account ID"
// @Success      200  {object}  models.BankAccountResponse  "Bank account retrieved successfully"
// @Response     400  {object}  object  "Bad request"
// @Response     404  {object}  object  "Bank account not found"
// @Router       /bank-account/{id} [get]
func (ctrl *bankAccountController) GetByID(c *gin.Context) {
	userID, err := getUserIDFromHeader(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid bank account ID",
		})
		return
	}

	response, err := ctrl.bankAccountSrv.GetByID(uint(id), userID)
	if err != nil {
		if err.Error() == "bank account not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"message": err.Error(),
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

// @Summary      Get all bank accounts
// @Description  Retrieves all bank accounts for the authenticated user
// @Tags         bank-accounts
// @Accept       json
// @Param        X-User-ID header string true "User ID"
// @Success      200  {array}  models.BankAccountResponse  "Bank accounts retrieved successfully"
// @Router       /bank-account [get]
func (ctrl *bankAccountController) GetAll(c *gin.Context) {
	userID, err := getUserIDFromHeader(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
		return
	}

	responses, err := ctrl.bankAccountSrv.GetByUserID(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, responses)
}

// @Summary      Get all bank accounts by user ID
// @Description  Retrieves all bank accounts for a specific user by user ID
// @Tags         bank-accounts
// @Accept       json
// @Param        userId path int true "User ID"
// @Success      200  {array}  models.BankAccountResponse  "Bank accounts retrieved successfully"
// @Response     400  {object}  object  "Bad request"
// @Router       /bank-account/user/{userId} [get]
func (ctrl *bankAccountController) GetByUserID(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid user ID",
		})
		return
	}

	accounts := []entities.BankAccount{
		{UserID: uint(userID), BankCode: "001", AccountNumber: "1234567890"},
		{UserID: uint(userID), BankCode: "002", AccountNumber: "0987654321"},
	}

	c.JSON(http.StatusOK, accounts)
}

// @Summary      Update a bank account
// @Description  Updates a specific bank account by ID for the authenticated user
// @Tags         bank-accounts
// @Accept       json
// @Param        X-User-ID header string true "User ID"
// @Param        id path int true "Bank Account ID"
// @Param        request body models.UpdateBankAccountRequest true "Bank account update details"
// @Success      200  {object}  models.BankAccountResponse  "Bank account updated successfully"
// @Response     400  {object}  object  "Bad request"
// @Response     404  {object}  object  "Bank account not found"
// @Router       /bank-account/{id} [put]
func (ctrl *bankAccountController) Update(c *gin.Context) {
	userID, err := getUserIDFromHeader(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid bank account ID",
		})
		return
	}

	var req models.UpdateBankAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request body or missing field: " + err.Error(),
		})
		return
	}

	response, err := ctrl.bankAccountSrv.Update(uint(id), userID, &req)
	if err != nil {
		if err.Error() == "bank account not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"message": err.Error(),
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

// @Summary      Delete a bank account
// @Description  Deletes a specific bank account by ID for the authenticated user
// @Tags         bank-accounts
// @Accept       json
// @Param        X-User-ID header string true "User ID"
// @Param        id path int true "Bank Account ID"
// @Success      200  {object}  nil  "Bank account deleted successfully"
// @Response     400  {object}  object  "Bad request"
// @Response     404  {object}  object  "Bank account not found"
// @Router       /bank-account/{id} [delete]
func (ctrl *bankAccountController) Delete(c *gin.Context) {
	userID, err := getUserIDFromHeader(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid bank account ID",
		})
		return
	}

	err = ctrl.bankAccountSrv.Delete(uint(id), userID)
	if err != nil {
		if err.Error() == "bank account not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"message": err.Error(),
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
		}
		return
	}

	c.Status(http.StatusOK)
}
