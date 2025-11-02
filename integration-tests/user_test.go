package integration_test

import (
	"banking-system/database"
	"banking-system/entities"
	"banking-system/models"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestRegister_Success(t *testing.T) {
	truncateTables()

	req := &models.RegisterRequest{
		Username: "johndoe",
		Password: "password123",
		Name:     "John Doe",
	}
	body, _ := json.Marshal(req)

	res := postRequest("/api/v1/user", body)

	assert.Equal(t, http.StatusCreated, res.Code)
	user := expectUserCreated(t, req.Username, req.Password, req.Name)
	expectWalletCreated(t, user.ID, 0.00)
}

func TestRegister_DuplicateUsername(t *testing.T) {
	truncateTables()

	firstReq, _ := json.Marshal(&models.RegisterRequest{
		Username: "johndoe",
		Password: "password123",
		Name:     "John Doe",
	})
	firstRes := postRequest("/api/v1/user", firstReq)
	assert.Equal(t, http.StatusCreated, firstRes.Code)

	secondReq, _ := json.Marshal(&models.RegisterRequest{
		Username: "johndoe",
		Password: "differentpassword",
		Name:     "Another John",
	})
	secondRes := postRequest("/api/v1/user", secondReq)

	assert.Equal(t, http.StatusBadRequest, secondRes.Code)
	expectUniqueUsername(t, "johndoe")
}

func TestRegister_MissingFields(t *testing.T) {
	truncateTables()

	tests := []struct {
		name    string
		request models.RegisterRequest
	}{
		{
			name: "missing username",
			request: models.RegisterRequest{
				Password: "password123",
				Name:     "John Doe",
			},
		},
		{
			name: "missing password",
			request: models.RegisterRequest{
				Username: "johndoe",
				Name:     "John Doe",
			},
		},
		{
			name: "missing name",
			request: models.RegisterRequest{
				Username: "johndoe",
				Password: "password123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := json.Marshal(&tt.request)
			res := postRequest("/api/v1/user", req)
			assert.Equal(t, http.StatusBadRequest, res.Code)
		})
	}
}

func expectUserCreated(t *testing.T, username, password, name string) entities.User {
	var user entities.User
	result := database.DB.Preload("Wallet").Where("username = ?", username).First(&user)

	assert.Nil(t, result.Error)
	assert.Equal(t, username, user.Username)
	assert.Equal(t, name, user.Name)

	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	assert.Nil(t, err, "Password should be hashed correctly")

	return user
}

func expectWalletCreated(t *testing.T, userID uint, amount float64) {
	var user entities.User
	result := database.DB.Preload("Wallet").First(&user, userID)

	assert.Nil(t, result.Error)
	assert.NotNil(t, user.Wallet)
	assert.Equal(t, 0.00, amount)
}

func expectUniqueUsername(t *testing.T, username string) {
	var count int64
	database.DB.Model(&entities.User{}).Where("username = ?", username).Count(&count)
	assert.Equal(t, int64(1), count)
}
