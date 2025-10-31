package controllers_test

import (
	"banking-system/router"
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDepositRoute(t *testing.T) {
	r := router.Setup()

	w := httptest.NewRecorder()
	jsonString := `{"user_id": 1, "currency": "TWD", "amount": 100, "payment_method": "XPay"}`
	req, _ := http.NewRequest("POST", "/payments/deposit", bytes.NewReader([]byte(jsonString)))
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
