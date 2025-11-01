package integration_test

import (
	"banking-system/database"
	"banking-system/entities"
	"banking-system/router"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var r *gin.Engine

func TestMain(m *testing.M) {
	r = router.Setup()
	database.ConnectTestDB()
	truncateTables()
	exitCode := m.Run()
	os.Exit(exitCode)
}

func truncateTables() {
	// Disable foreign key checks
	database.DB.Exec("SET session_replication_role = 'replica';")

	var tables = []string{
		"users",
		"wallets",
		"bank_accounts",
		"transactions",
	}

	for _, tableName := range tables {
		if err := database.DB.Exec("TRUNCATE TABLE " + tableName + " RESTART IDENTITY CASCADE;").Error; err != nil {
			log.Fatalf("Failed to truncate table %s: %v", tableName, err)
		}
	}

	// Re-enable foreign key checks
	database.DB.Exec("SET session_replication_role = 'origin';")
}

func TestValidDeposit(t *testing.T) {
	user := givenUserBalance(0)

	input := fmt.Sprintf(`{"user_id": %d, "currency": "%s", "amount": %.2f, "payment_method": "%s"}`, user.ID, user.Wallet.Currency, 100.00, "AnyPay")
	w := httptest.NewRecorder()
	postRequest(w, input)

	assert.Equal(t, 200, w.Code)
	expectTransactionEqual(t, &entities.Transaction{
		WalletID:      user.Wallet.ID,
		Type:          entities.TransactionTypes.Deposit,
		Status:        entities.TransactionStatuses.Pending,
		Amount:        100.00,
		PaymentMethod: "AnyPay",
	})
}

func postRequest(res *httptest.ResponseRecorder, body string) {
	req, _ := http.NewRequest("POST", "/payments/deposit", bytes.NewReader([]byte(body)))
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(res, req)
}

func givenUserBalance(amount float64) *entities.User {
	user := &entities.User{
		Username:     "test_user",
		PasswordHash: "any",
		Wallet: entities.Wallet{
			Balance:  amount,
			Currency: "TWD",
		},
	}
	database.DB.Create(user)
	return user
}

func expectTransactionEqual(t *testing.T, expected *entities.Transaction) {
	var tx entities.Transaction
	result := database.DB.Where("wallet_id = ?", expected.WalletID).First(&tx)

	assert.Nil(t, result.Error)
	assert.Equal(t, expected.Status, tx.Status)
	assert.Equal(t, expected.Type, tx.Type)
	assert.Equal(t, expected.Amount, tx.Amount)
	assert.Equal(t, expected.PaymentMethod, tx.PaymentMethod)
}
