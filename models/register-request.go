package models

type RegisterRequest struct {
	Username string `json:"username" binding:"required,max=20"`
	Password string `json:"password" binding:"required,max=20"`
	Name     string `json:"name" binding:"required,max=100"`
}
