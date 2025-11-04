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
	"fmt"
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
	pspFactoryMock      *pspMock.MockPSPFactory
	paymentProviderMock *pspMock.MockPaymentServiceProvider
)

func TestDeposit(t *testing.T) {
	truncateTables()
	ctrl := gomock.NewController(t)
	pspFactoryMock = pspMock.NewMockPSPFactory(ctrl)
	paymentProviderMock = pspMock.NewMockPaymentServiceProvider(ctrl)
	sut := controllers.NewPaymentController(services.NewPaymentService(repos.NewUserRepo(), repos.NewTransactionRepo(), pspFactoryMock))

	txUUID := uuid.New()
	const redirectUrl = "https://external.payment.page/payin"
	givenPayInResponse(txUUID.String(), redirectUrl)
	user := givenUserHasBalance(0)

	req, _ := json.Marshal(&models.DepositRequest{
		UUID:          txUUID,
		Amount:        100.00,
		PaymentMethod: "AnyPay",
	})
	res := postRequestWithHandler("/api/v1/payments/deposit", sut.Deposit, req, user.ID)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, redirectUrl, getResponseField(res, "redirect_url"))
	expectTransactionEqual(t, &entities.Transaction{
		UUID:          txUUID,
		WalletID:      user.Wallet.ID,
		Type:          entities.TransactionTypes.Deposit,
		Status:        entities.TransactionStatuses.Pending,
		Amount:        100.00,
		PaymentMethod: "AnyPay",
	})
}

func TestDeposit_DuplicateRequests(t *testing.T) {
	truncateTables()
	ctrl := gomock.NewController(t)
	pspFactoryMock = pspMock.NewMockPSPFactory(ctrl)
	paymentProviderMock = pspMock.NewMockPaymentServiceProvider(ctrl)

	// // assert PayIn is called only once
	expectPayInCalledOnce()

	sut := controllers.NewPaymentController(services.NewPaymentService(repos.NewUserRepo(), repos.NewTransactionRepo(), pspFactoryMock))

	user := givenUserHasBalance(0)

	txUUID := uuid.New()
	req, _ := json.Marshal(&models.DepositRequest{
		UUID:          txUUID,
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

			res := postRequestWithHandler("/api/v1/payments/deposit", sut.Deposit, req, user.ID)

			if res.Code == http.StatusOK {
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
		TransactionID: uuid.NewString(),
	}

	user := givenUserHasBalance(100)
	givenTransaction(&entities.Transaction{
		UUID:     uuid.MustParse(req.TransactionID),
		Type:     entities.TransactionTypes.Deposit,
		Status:   entities.TransactionStatuses.Pending,
		Amount:   50.00,
		WalletID: user.Wallet.ID,
	})

	body, _ := json.Marshal(req)
	res := postRequestWithPSPAuth("/api/v1/payments/confirm", body)

	assert.Equal(t, http.StatusOK, res.Code)
	expectTransactionStatus(t, req.TransactionID, entities.TransactionStatuses.Completed)
	expectBalance(t, user.Wallet.ID, 150.00)
}

func TestDepositConfirm_DuplicateRequest(t *testing.T) {
	truncateTables()

	req := &psp.ConfirmRequest{
		TransactionID: uuid.NewString(),
	}

	user := givenUserHasBalance(100)
	givenTransaction(&entities.Transaction{
		UUID:     uuid.MustParse(req.TransactionID),
		Type:     entities.TransactionTypes.Deposit,
		Status:   entities.TransactionStatuses.Pending,
		Amount:   50.00,
		WalletID: user.Wallet.ID,
	})

	body, _ := json.Marshal(req)
	firstResp := postRequestWithPSPAuth("/api/v1/payments/confirm", body)
	assert.Equal(t, http.StatusOK, firstResp.Code)
	expectTransactionStatus(t, req.TransactionID, entities.TransactionStatuses.Completed)
	expectBalance(t, user.Wallet.ID, 150.00)

	secondResp := postRequestWithPSPAuth("/api/v1/payments/confirm", body)
	assert.Equal(t, http.StatusOK, secondResp.Code)
	expectTransactionStatus(t, req.TransactionID, entities.TransactionStatuses.Completed)
	expectBalance(t, user.Wallet.ID, 150.00)
}

func TestDepositConfirm_ConcurrentRequests(t *testing.T) {
	truncateTables()

	req := &psp.ConfirmRequest{
		TransactionID: uuid.NewString(),
	}

	user := givenUserHasBalance(100)
	givenTransaction(&entities.Transaction{
		UUID:     uuid.MustParse(req.TransactionID),
		Type:     entities.TransactionTypes.Deposit,
		Status:   entities.TransactionStatuses.Pending,
		Amount:   50.00,
		WalletID: user.Wallet.ID,
	})

	// Simulate 10 concurrent requests
	concurrentRequests := 10
	var wg sync.WaitGroup
	wg.Add(concurrentRequests)

	body, _ := json.Marshal(req)
	for i := range concurrentRequests {
		go func(index int) {
			defer wg.Done()
			postRequestWithPSPAuth("/api/v1/payments/confirm", body)
		}(i)
	}

	wg.Wait()

	expectTransactionStatus(t, req.TransactionID, entities.TransactionStatuses.Completed)
	expectBalance(t, user.Wallet.ID, 150.00)
}

func TestDepositCancel(t *testing.T) {
	truncateTables()

	req := &psp.CancelRequest{
		TransactionID: "a05aa863-d9ab-42e6-8122-f76e43edaa22",
	}

	user := givenUserHasBalance(100)
	givenTransaction(&entities.Transaction{
		UUID:     uuid.MustParse(req.TransactionID),
		Type:     entities.TransactionTypes.Deposit,
		Status:   entities.TransactionStatuses.Pending,
		Amount:   50.00,
		WalletID: user.Wallet.ID,
	})

	body, _ := json.Marshal(req)
	res := postRequestWithPSPAuth("/api/v1/payments/cancel", body)

	assert.Equal(t, http.StatusOK, res.Code)
	expectTransactionStatus(t, req.TransactionID, entities.TransactionStatuses.Canceled)
	expectBalance(t, user.Wallet.ID, 100.00) // Balance should not change
}

func TestDepositCancel_DuplicateRequest(t *testing.T) {
	truncateTables()

	req := &psp.CancelRequest{
		TransactionID: "b05aa863-d9ab-42e6-8122-f76e43edaa23",
	}

	user := givenUserHasBalance(100)
	givenTransaction(&entities.Transaction{
		UUID:     uuid.MustParse(req.TransactionID),
		Type:     entities.TransactionTypes.Deposit,
		Status:   entities.TransactionStatuses.Pending,
		Amount:   50.00,
		WalletID: user.Wallet.ID,
	})

	body, _ := json.Marshal(req)
	firstResp := postRequestWithPSPAuth("/api/v1/payments/cancel", body)
	assert.Equal(t, http.StatusOK, firstResp.Code)
	expectTransactionStatus(t, req.TransactionID, entities.TransactionStatuses.Canceled)
	expectBalance(t, user.Wallet.ID, 100.00)

	secondResp := postRequestWithPSPAuth("/api/v1/payments/cancel", body)
	assert.Equal(t, http.StatusOK, secondResp.Code)
	expectTransactionStatus(t, req.TransactionID, entities.TransactionStatuses.Canceled)
	expectBalance(t, user.Wallet.ID, 100.00) // Balance should still be unchanged
}

func TestWithdraw_Success(t *testing.T) {
	truncateTables()
	ctrl := gomock.NewController(t)
	pspFactoryMock = pspMock.NewMockPSPFactory(ctrl)
	paymentProviderMock = pspMock.NewMockPaymentServiceProvider(ctrl)
	sut := controllers.NewPaymentController(services.NewPaymentService(repos.NewUserRepo(), repos.NewTransactionRepo(), pspFactoryMock))

	txUUID := uuid.New()
	givenPayOutResponse(txUUID.String())
	user := givenUserHasBalance(200.00)

	req, _ := json.Marshal(&models.WithdrawRequest{
		UUID:          txUUID,
		PaymentMethod: "AnyPay",
		Amount:        50.00,
	})
	res := postRequestWithHandler("/api/v1/payments/withdraw", sut.Withdraw, req, user.ID)

	assert.Equal(t, http.StatusOK, res.Code)
	expectTransactionEqual(t, &entities.Transaction{
		UUID:          txUUID,
		WalletID:      user.Wallet.ID,
		Type:          entities.TransactionTypes.Withdrawal,
		Status:        entities.TransactionStatuses.Pending,
		PaymentMethod: "AnyPay",
		Amount:        50.00,
	})
	expectBalance(t, user.Wallet.ID, 150.00) // Balance should be deducted
}

func TestWithdraw_InsufficientBalance(t *testing.T) {
	truncateTables()
	ctrl := gomock.NewController(t)
	pspFactoryMock = pspMock.NewMockPSPFactory(ctrl)
	paymentProviderMock = pspMock.NewMockPaymentServiceProvider(ctrl)
	sut := controllers.NewPaymentController(services.NewPaymentService(repos.NewUserRepo(), repos.NewTransactionRepo(), pspFactoryMock))

	txUUID := uuid.New()
	user := givenUserHasBalance(30.00)

	req, _ := json.Marshal(&models.WithdrawRequest{
		UUID:          txUUID,
		PaymentMethod: "AnyPay",
		Amount:        50.00, // More than available balance
	})
	res := postRequestWithHandler("/api/v1/payments/withdraw", sut.Withdraw, req, user.ID)

	assert.Equal(t, http.StatusBadRequest, res.Code)
	expectBalance(t, user.Wallet.ID, 30.00) // Balance should remain unchanged
}

func TestWithdraw_DuplicateRequests(t *testing.T) {
	truncateTables()
	ctrl := gomock.NewController(t)
	pspFactoryMock = pspMock.NewMockPSPFactory(ctrl)
	paymentProviderMock = pspMock.NewMockPaymentServiceProvider(ctrl)

	user := givenUserHasBalance(200.00)

	txUUID := uuid.New()
	req, _ := json.Marshal(&models.WithdrawRequest{
		UUID:          txUUID,
		PaymentMethod: "AnyPay",
		Amount:        50.00,
	})

	// assert PayOut is called only once
	expectPayOutCalledOnce()

	sut := controllers.NewPaymentController(services.NewPaymentService(repos.NewUserRepo(), repos.NewTransactionRepo(), pspFactoryMock))

	// Simulate 10 concurrent requests
	concurrentRequests := 10
	successCount := make(chan bool, concurrentRequests)
	failureCount := make(chan bool, concurrentRequests)
	var wg sync.WaitGroup
	wg.Add(concurrentRequests)

	for i := range concurrentRequests {
		go func(index int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					failureCount <- true
					return
				}
			}()

			res := postRequestWithHandler("/api/v1/payments/withdraw", sut.Withdraw, req, user.ID)

			if res.Code == http.StatusOK {
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
	expectBalance(t, user.Wallet.ID, 150.00) // Balance should be deducted only once
}

func TestWithdrawCancel(t *testing.T) {
	truncateTables()

	req := &psp.CancelRequest{
		TransactionID: uuid.NewString(),
	}

	user := givenUserHasBalance(100.00)
	givenTransaction(&entities.Transaction{
		UUID:     uuid.MustParse(req.TransactionID),
		Type:     entities.TransactionTypes.Withdrawal,
		Status:   entities.TransactionStatuses.Pending,
		Amount:   10.00,
		WalletID: user.Wallet.ID,
	})

	body, _ := json.Marshal(req)
	res := postRequestWithPSPAuth("/api/v1/payments/cancel", body)

	assert.Equal(t, http.StatusOK, res.Code)
	expectTransactionStatus(t, req.TransactionID, entities.TransactionStatuses.Canceled)
	expectBalance(t, user.Wallet.ID, 110.00, "Balance should be refunded back")
}

func TestWithdrawCancel_ConcurrentRequests(t *testing.T) {
	truncateTables()

	req := &psp.CancelRequest{
		TransactionID: uuid.NewString(),
	}
	body, _ := json.Marshal(req)

	user := givenUserHasBalance(100.00)
	givenTransaction(&entities.Transaction{
		UUID:     uuid.MustParse(req.TransactionID),
		Type:     entities.TransactionTypes.Withdrawal,
		Status:   entities.TransactionStatuses.Pending,
		Amount:   10.00,
		WalletID: user.Wallet.ID,
	})

	// Simulate 10 concurrent requests
	concurrentRequests := 10
	var wg sync.WaitGroup
	wg.Add(concurrentRequests)
	successCount := make(chan bool, concurrentRequests)
	failureCount := make(chan bool, concurrentRequests)

	for i := range concurrentRequests {
		go func(index int) {
			defer wg.Done()

			res := postRequestWithPSPAuth("/api/v1/payments/cancel", body)

			if res.Code == http.StatusOK {
				successCount <- true
			} else {
				failureCount <- true
			}
		}(i)
	}

	wg.Wait()
	close(successCount)
	close(failureCount)

	assert.Equal(t, concurrentRequests, len(successCount), "All requests should succeed.")
	expectTransactionStatus(t, req.TransactionID, entities.TransactionStatuses.Canceled)
	expectBalance(t, user.Wallet.ID, 110.00, "Balance should be refunded only once")
}

func TestTransfer_Success(t *testing.T) {
	truncateTables()

	sender := givenUserHasBalance(200.00)
	recipient := givenUserHasBalance(50.00)

	transferOutUUID := uuid.New()
	req, _ := json.Marshal(&models.TransferRequest{
		UUID:            transferOutUUID,
		SenderUserID:    sender.ID,
		RecipientUserID: recipient.ID,
		Amount:          10.00,
	})

	res := postRequest("/api/v1/payments/transfer", req)

	assert.Equal(t, http.StatusOK, res.Code)
	expectBalance(t, sender.Wallet.ID, 190.00)
	expectBalance(t, recipient.Wallet.ID, 60.00)
	expectTransferTransactionLinked(t, transferOutUUID)
}

func TestTransfer_InsufficientBalance(t *testing.T) {
	truncateTables()

	sender := givenUserHasBalance(30.00)
	recipient := givenUserHasBalance(50.00)

	txUUID := uuid.New()
	req, _ := json.Marshal(&models.TransferRequest{
		UUID:            txUUID,
		SenderUserID:    sender.ID,
		RecipientUserID: recipient.ID,
		Amount:          100.00, // More than sender's balance
	})

	res := postRequest("/api/v1/payments/transfer", req)

	assert.Equal(t, http.StatusBadRequest, res.Code)
	expectBalance(t, sender.Wallet.ID, 30.00, "Balance should remain unchanged")
	expectBalance(t, recipient.Wallet.ID, 50.00, "Balance should remain unchanged")
}

func TestTransfer_SameUser(t *testing.T) {
	truncateTables()

	user := givenUserHasBalance(100.00)

	txUUID := uuid.New()
	req, _ := json.Marshal(&models.TransferRequest{
		UUID:            txUUID,
		SenderUserID:    user.ID,
		RecipientUserID: user.ID, // Same user
		Amount:          50.00,
	})

	res := postRequest("/api/v1/payments/transfer", req)

	assert.Equal(t, http.StatusBadRequest, res.Code)
	expectBalance(t, user.Wallet.ID, 100.00, "Balance should remain unchanged")
}

func TestTransfer_BelowMinimum(t *testing.T) {
	truncateTables()

	sender := givenUserHasBalance(100.00)
	recipient := givenUserHasBalance(50.00)

	txUUID := uuid.New()
	req, _ := json.Marshal(&models.TransferRequest{
		UUID:            txUUID,
		SenderUserID:    sender.ID,
		RecipientUserID: recipient.ID,
		Amount:          0.50, // Below minimum
	})

	res := postRequest("/api/v1/payments/transfer", req)

	assert.Equal(t, http.StatusBadRequest, res.Code)
	expectBalance(t, sender.Wallet.ID, 100.00, "Balance should remain unchanged")
	expectBalance(t, recipient.Wallet.ID, 50.00, "Balance should remain unchanged")
}

func TestTransfer_AboveMaximum(t *testing.T) {
	truncateTables()

	sender := givenUserHasBalance(200000.00)
	recipient := givenUserHasBalance(50.00)

	txUUID := uuid.New()
	req, _ := json.Marshal(&models.TransferRequest{
		UUID:            txUUID,
		SenderUserID:    sender.ID,
		RecipientUserID: recipient.ID,
		Amount:          150000.00, // Above maximum
	})

	res := postRequest("/api/v1/payments/transfer", req)

	assert.Equal(t, http.StatusBadRequest, res.Code)
	expectBalance(t, sender.Wallet.ID, 200000.00, "Balance should remain unchanged")
	expectBalance(t, recipient.Wallet.ID, 50.00, "Balance should remain unchanged")
}

func TestTransfer_ConcurrentRequests(t *testing.T) {
	truncateTables()

	sender := givenUserHasBalance(1000.00)
	recipient := givenUserHasBalance(50.00)

	transferOutUUID := uuid.New()
	transferRequest := &models.TransferRequest{
		UUID:            transferOutUUID,
		SenderUserID:    sender.ID,
		RecipientUserID: recipient.ID,
		Amount:          100.00,
	}

	// Simulate 10 concurrent transfer requests with the same UUID
	concurrentRequests := 10
	var wg sync.WaitGroup
	wg.Add(concurrentRequests)

	body, _ := json.Marshal(transferRequest)
	for i := range concurrentRequests {
		go func(index int) {
			defer wg.Done()
			postRequest("/api/v1/payments/transfer", body)
		}(i)
	}

	wg.Wait()

	expectBalance(t, sender.Wallet.ID, 900.00, "Sender balance should be deducted only once")
	expectBalance(t, recipient.Wallet.ID, 150.00, "Recipient balance should be credited only once")
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

func postRequestWithPSPAuth(path string, body []byte) *httptest.ResponseRecorder {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", path, bytes.NewReader(body))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-PSP-API-Key", "psp_secret_key_12345")
	r.ServeHTTP(res, req)
	return res
}

func postRequestWithHandler(path string, handler func(c *gin.Context), body []byte, userID ...uint) *httptest.ResponseRecorder {
	res := httptest.NewRecorder()
	ctx, r := gin.CreateTestContext(res)
	r.POST(path, handler)
	ctx.Request = httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")

	// Set X-USER-ID header if provided
	if len(userID) > 0 {
		ctx.Request.Header.Set("X-User-ID", fmt.Sprintf("%d", userID[0]))
	}

	r.ServeHTTP(res, ctx.Request)
	return res
}

func givenUserHasBalance(amount float64) *entities.User {
	user := &entities.User{
		Username:     "usr_" + uuid.NewString()[:8],
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
	pspFactoryMock.EXPECT().
		NewPaymentServiceProvider(gomock.Any()).
		Return(paymentProviderMock)

	paymentProviderMock.EXPECT().PayIn(gomock.Any()).
		Return(&psp.PayInResponse{
			TransactionID: txUUID,
			RedirectUrl:   redirectUrl,
		}, nil)
}

func expectPayInCalledOnce() {
	pspFactoryMock.EXPECT().
		NewPaymentServiceProvider(gomock.Any()).
		Return(paymentProviderMock).
		Times(1)

	paymentProviderMock.EXPECT().PayIn(gomock.Any()).
		Return(&psp.PayInResponse{}, nil).
		Times(1)
}

func givenPayOutResponse(txID string) {
	pspFactoryMock.EXPECT().
		NewPaymentServiceProvider(gomock.Any()).
		Return(paymentProviderMock).
		Times(1)

	paymentProviderMock.EXPECT().PayOut().
		Return(&psp.PayOutResponse{
			TransactionID: txID,
		}, nil).
		Times(1)
}

func expectPayOutCalledOnce() {
	pspFactoryMock.EXPECT().
		NewPaymentServiceProvider(gomock.Any()).
		Return(paymentProviderMock).
		Times(1)

	paymentProviderMock.EXPECT().PayOut().
		Return(&psp.PayOutResponse{}, nil).
		Times(1)
}

func givenTransaction(transaction *entities.Transaction) {
	database.DB.Create(transaction)
}

func expectTransactionEqual(t *testing.T, expected *entities.Transaction) {
	var actual entities.Transaction
	result := database.DB.First(&actual, expected.UUID)

	assert.Nil(t, result.Error)
	assert.Equal(t, expected.WalletID, actual.WalletID)
	assert.Equal(t, expected.Status, actual.Status)
	assert.Equal(t, expected.Type, actual.Type)
	assert.Equal(t, expected.Amount, actual.Amount)
	assert.Equal(t, expected.PaymentMethod, actual.PaymentMethod)
}

func expectTransferTransactionLinked(t *testing.T, transferOutUUID uuid.UUID) {
	var actual entities.Transaction
	result := database.DB.Preload("RelatedTransaction").First(&actual, transferOutUUID)

	assert.Nil(t, result.Error)
	assert.NotNil(t, actual.RelatedTransaction)
	assert.Equal(t, actual.UUID.String(), actual.RelatedTransaction.RelatedTransactionID.String())
	assert.Equal(t, actual.Status, actual.RelatedTransaction.Status)
	assert.Equal(t, actual.Amount, actual.RelatedTransaction.Amount)
	assert.Equal(t, entities.TransactionTypes.TransferOut, actual.Type)
	assert.Equal(t, entities.TransactionTypes.TransferIn, actual.RelatedTransaction.Type)
}

func expectTransactionStatus(t *testing.T, transactionId string, transactionStatus entities.TransactionStatus) {
	var tx entities.Transaction
	result := database.DB.Where("uuid = ?", transactionId).First(&tx)

	assert.Nil(t, result.Error)
	assert.Equal(t, transactionStatus, tx.Status)
}

func expectBalance(t *testing.T, walletId uint, amount float64, msgAndArgs ...interface{}) {
	var wallet entities.Wallet
	result := database.DB.Where("id = ?", walletId).First(&wallet)

	assert.Nil(t, result.Error)
	assert.Equal(t, amount, wallet.Balance, msgAndArgs...)
}

func getResponseField(resp *httptest.ResponseRecorder, field string) string {
	var response map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &response)
	return response[field].(string)
}
