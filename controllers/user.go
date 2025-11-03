package controllers

import (
	"banking-system/entities"
	"banking-system/models"
	"banking-system/services"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type UserController interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
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

// @Summary      User Login
// @Description  Authenticates a user with username and password, and sets an authorization cookie
// @Tags         users
// @Accept       json
// @Param        request body models.LoginRequest true "Login form data"
// @Response     200  	  {object}  nil  "User logged in successfully"
// @Failure      400      {object}  object  "Bad Request (e.g., invalid body or validation error)"
// @Failure      401      {object}  object  "Unauthorized (Invalid username or password)"
// @Router       /user/login [post]
func (ctrl *userController) Login(c *gin.Context) {
	// Retrieve req from request
	var req models.LoginRequest
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// validate input
	if !ctrl.validateRequest(&req) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid username or password"})
		return
	}

	// Validate username
	valid, foundUser := ctrl.validateExistence(req.Username)
	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// validate password
	if !ctrl.validatePassword(&req, foundUser.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Generate JWT token
	tokenString := ctrl.generateToken(foundUser)

	// Write token to response cookie
	c.SetSameSite(http.SameSiteNoneMode)
	c.SetCookie("authorization", tokenString, 60*60*24*7, "/", "", true, true)
	c.Status(http.StatusOK)
}

func (*userController) validateRequest(req any) bool {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return validate.Struct(req) == nil
}

func (ctrl *userController) validateExistence(username string) (bool, *entities.User) {
	foundUser, err := ctrl.userSrv.GetByUsername(username)
	if err != nil {
		return false, nil
	}

	return foundUser != nil, foundUser
}

func (*userController) validatePassword(req *models.LoginRequest, correctPassword string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(correctPassword), []byte(req.Password)); err != nil {
		return false
	}
	return true
}

func (*userController) generateToken(user *entities.User) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 72).Unix(),
	})

	key := os.Getenv("USER_TOKEN_SECRET_KEY")
	if key == "" {
		log.Panic("USER_TOKEN_SECRET_KEY environment variable not set")
	}

	tokenString, err := token.SignedString([]byte(key))
	if err != nil {
		log.Panicf("Failed to sign token: %v", err)
	}

	return tokenString
}
