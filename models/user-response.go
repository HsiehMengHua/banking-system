package models

type UserInfoResponse struct {
	Username string  `json:"username"`
	Balance  float64 `json:"balance"`
}
