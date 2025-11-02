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
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	pspMock "banking-system/psp/mock"
)

var r *gin.Engine

func TestMain(m *testing.M) {
	r = router.Setup()
	database.ConnectTestDB()
	exitCode := m.Run()
	os.Exit(exitCode)
}

var (
	paymentServiceProviderMock *pspMock.MockPaymentServiceProvider
)

func TestDeposit(t *testing.T) {
	truncateTables()
	ctrl := gomock.NewController(t)
	paymentServiceProviderMock = pspMock.NewMockPaymentServiceProvider(ctrl)
	sut := controllers.NewPaymentController(services.NewPaymentService(repos.NewUserRepo(), repos.NewTransactionRepo(), paymentServiceProviderMock))

	txUUID := uuid.New()
	const redirectUrl = "https://external.payment.page/payin"
	givenPayInResponse(txUUID.String(), redirectUrl)
	user := givenUserHasBalance(0)

	req, _ := json.Marshal(&models.DepositRequest{
		UUID:          txUUID,
		UserID:        user.ID,
		Currency:      user.Wallet.Currency,
		Amount:        100.00,
		PaymentMethod: "AnyPay",
	})
	res := postRequestWithHandler("/api/v1/payments/deposit", sut.Deposit, req)

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

func TestDeposit_DuplicateRequests(t *testing.T) {
	truncateTables()
	ctrl := gomock.NewController(t)
	paymentServiceProviderMock = pspMock.NewMockPaymentServiceProvider(ctrl)
	paymentServiceProviderMock.EXPECT().PayIn().Return(&psp.DepositResponse{}, nil).Times(1)

	sut := controllers.NewPaymentController(services.NewPaymentService(repos.NewUserRepo(), repos.NewTransactionRepo(), paymentServiceProviderMock))

	user := givenUserHasBalance(0)

	txUUID := uuid.New()
	req, _ := json.Marshal(&models.DepositRequest{
		UUID:          txUUID,
		UserID:        user.ID,
		Currency:      user.Wallet.Currency,
		Amount:        100.00,
		PaymentMethod: "AnyPay",
	})

	// Simulate 10 concurrent requests
	concurrentRequests := 10
	successCount := make(chan bool, concurrentRequests)
	failureCount := make(chan bool, concurrentRequests)
	var wg sync.WaitGroup
	wg.Add(concurrentRequests)

	for i := 0; i < concurrentRequests; i++ {
		go func(index int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					failureCount <- true
					return
				}
			}()

			res := postRequestWithHandler("/api/v1/payments/deposit", sut.Deposit, req)

			if res.Code == http.StatusFound {
				successCount <- true
			} else {
				failureCount <- true
			}
		}(i)
	}

	wg.Wait()
	close(successCount)
	close(failureCount)

	assert.Equal(t, 1, len(successCount), "Exactly one request should succeed.")
	assert.Equal(t, concurrentRequests-1, len(failureCount), "The remaining requests should have failed.")
}

func TestDepositConfirm(t *testing.T) {
	truncateTables()
	req := &psp.ConfirmRequest{
		TransactionID: "705aa863-d9ab-42e6-8122-f76e43edaa19",
		Amount:        50.00,
	}

	user := givenUserHasBalance(100)
	givenTransaction(&entities.Transaction{
		UUID:     uuid.MustParse(req.TransactionID),
		Type:     entities.TransactionTypes.Deposit,
		Status:   entities.TransactionStatuses.Pending,
		Amount:   req.Amount,
		WalletID: user.Wallet.ID,
	})

	body, _ := json.Marshal(req)
	res := postRequest("/api/v1/payments/confirm", body)

	assert.Equal(t, http.StatusOK, res.Code)
	expectTransactionStatus(t, req.TransactionID, entities.TransactionStatuses.Completed)
	expectBalance(t, user.Wallet.ID, 150.00)
}

func TestDepositConfirm_DuplicateRequest(t *testing.T) {
	truncateTables()

	req := &psp.ConfirmRequest{
		TransactionID: "805aa863-d9ab-42e6-8122-f76e43edaa20",
		Amount:        50.00,
	}

	user := givenUserHasBalance(100)
	givenTransaction(&entities.Transaction{
		UUID:     uuid.MustParse(req.TransactionID),
		Type:     entities.TransactionTypes.Deposit,
		Status:   entities.TransactionStatuses.Pending,
		Amount:   req.Amount,
		WalletID: user.Wallet.ID,
	})

	body, _ := json.Marshal(req)
	firstResp := postRequest("/api/v1/payments/confirm", body)
	assert.Equal(t, http.StatusOK, firstResp.Code)
	expectTransactionStatus(t, req.TransactionID, entities.TransactionStatuses.Completed)
	expectBalance(t, user.Wallet.ID, 150.00)

	secondResp := postRequest("/api/v1/payments/confirm", body)
	assert.Equal(t, http.StatusOK, secondResp.Code)
	expectTransactionStatus(t, req.TransactionID, entities.TransactionStatuses.Completed)
	expectBalance(t, user.Wallet.ID, 150.00)
}

func TestDepositConfirm_ConcurrentRequests(t *testing.T) {
	truncateTables()

	req := &psp.ConfirmRequest{
		TransactionID: "905aa863-d9ab-42e6-8122-f76e43edaa21",
		Amount:        50.00,
	}

	user := givenUserHasBalance(100)
	givenTransaction(&entities.Transaction{
		UUID:     uuid.MustParse(req.TransactionID),
		Type:     entities.TransactionTypes.Deposit,
		Status:   entities.TransactionStatuses.Pending,
		Amount:   req.Amount,
		WalletID: user.Wallet.ID,
	})

	// Simulate 10 concurrent requests
	concurrentRequests := 10
	var wg sync.WaitGroup
	wg.Add(concurrentRequests)

	body, _ := json.Marshal(req)
	for i := 0; i < concurrentRequests; i++ {
		go func(index int) {
			defer wg.Done()
			postRequest("/api/v1/payments/confirm", body)
		}(i)
	}

	wg.Wait()

	expectTransactionStatus(t, req.TransactionID, entities.TransactionStatuses.Completed)
	expectBalance(t, user.Wallet.ID, 150.00)
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

func postRequest(path string, body []byte) *httptest.ResponseRecorder {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", path, bytes.NewReader(body))
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(res, req)
	return res
}

func postRequestWithHandler(path string, handler func(c *gin.Context), body []byte) *httptest.ResponseRecorder {
	res := httptest.NewRecorder()
	ctx, r := gin.CreateTestContext(res)
	r.POST(path, handler)
	ctx.Request = httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(res, ctx.Request)
	return res
}

func givenUserHasBalance(amount float64) *entities.User {
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

func givenPayInResponse(txUUID string, redirectUrl string) {
	paymentServiceProviderMock.EXPECT().PayIn().
		Return(&psp.DepositResponse{
			TransactionID: txUUID,
			RedirectUrl:   redirectUrl,
		}, nil).
		Times(1)
}

func givenTransaction(transaction *entities.Transaction) {
	database.DB.Create(transaction)
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

func expectTransactionStatus(t *testing.T, transactionId string, transactionStatus entities.TransactionStatus) {
	var tx entities.Transaction
	result := database.DB.Where("uuid = ?", transactionId).First(&tx)

	assert.Nil(t, result.Error)
	assert.Equal(t, transactionStatus, tx.Status)
}

func expectBalance(t *testing.T, walletId uint, amount float64) {
	var wallet entities.Wallet
	result := database.DB.Where("id = ?", walletId).First(&wallet)

	assert.Nil(t, result.Error)
	assert.Equal(t, amount, wallet.Balance)
}
