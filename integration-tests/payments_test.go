package integration_test

import (
	"banking-system/controllers"
	"banking-system/database"
	"banking-system/entities"
	"banking-system/models"
	"banking-system/psp"
	"banking-system/repos"
	"banking-system/router"
	"banking-system/services"
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	pspMock "banking-system/psp/mock"
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

var (
	paymentServiceProviderMock *pspMock.MockPaymentServiceProvider
)

func TestValidDeposit(t *testing.T) {
	ctrl := gomock.NewController(t)
	paymentServiceProviderMock = pspMock.NewMockPaymentServiceProvider(ctrl)
	sut := controllers.NewPaymentController(services.NewPaymentService(repos.NewUserRepo(), repos.NewTransactionRepo(), paymentServiceProviderMock))

	const redirectUrl = "https://external.payment.page/payin"
	givenDepositRedirectUrl(redirectUrl)
	user := givenUserBalance(0)

	req, _ := json.Marshal(&models.DepositRequest{
		UserID:        user.ID,
		Currency:      user.Wallet.Currency,
		Amount:        100.00,
		PaymentMethod: "AnyPay",
	})
	res := postRequest("/payments/deposit", sut.Deposit, string(req))

	expectTransactionEqual(t, &entities.Transaction{
		WalletID:      user.Wallet.ID,
		Type:          entities.TransactionTypes.Deposit,
		Status:        entities.TransactionStatuses.Pending,
		Amount:        100.00,
		PaymentMethod: "AnyPay",
	})
	assert.Equal(t, http.StatusFound, res.Code)
	assert.Equal(t, redirectUrl, res.Result().Header.Get("Location"))
}

func postRequest(path string, handler func(c *gin.Context), body string) *httptest.ResponseRecorder {
	res := httptest.NewRecorder()
	ctx, r := gin.CreateTestContext(res)
	r.POST(path, handler)
	ctx.Request = httptest.NewRequest(http.MethodPost, path, bytes.NewReader([]byte(body)))
	ctx.Request.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(res, ctx.Request)
	return res
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

func givenDepositRedirectUrl(redirectUrl string) {
	paymentServiceProviderMock.EXPECT().PayIn().
		Return(&psp.DepositResponse{
			TransactionID: "tx_123456",
			RedirectUrl:   redirectUrl,
		}, nil).
		Times(1)
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
