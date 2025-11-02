package controllers

import (
	"banking-system/models"
	"banking-system/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserController interface {
	Register(c *gin.Context)
}

type userController struct {
	userSrv services.UserService
}

func NewUserController(userSrv services.UserService) UserController {
	return &userController{
		userSrv: userSrv,
	}
}

// @Summary      Register a new user
// @Description  Creates a new user with hashed password and an associated wallet with initial balance of 0
// @Tags         users
// @Accept       json
// @Param        request body models.RegisterRequest true "User registration details"
// @Response     201  {object}  nil  "User created successfully"
// @Response     400  {object}  object  "Bad request - validation error or user already exists"
// @Router       /user [post]
func (ctrl *userController) Register(c *gin.Context) {
	var req models.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request body or missing field: " + err.Error(),
		})
		return
	}

	if err := ctrl.userSrv.Register(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.Status(http.StatusCreated)
}
