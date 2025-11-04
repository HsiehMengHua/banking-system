package repos

import (
	"banking-system/database"
	"banking-system/entities"
)

//go:generate mockgen -source=userRepo.go -destination=mock/userRepo.go

type UserRepo interface {
	Create(user *entities.User) error
	Get(id uint) (*entities.User, error)
	GetByUsername(username string) (*entities.User, error)
	UpdateWallet(user *entities.User) error
}

type userRepo struct {
}

func NewUserRepo() UserRepo {
	return &userRepo{}
}

func (*userRepo) Create(user *entities.User) error {
	result := database.DB.Create(user)
	return result.Error
}

func (*userRepo) Get(id uint) (*entities.User, error) {
	var user entities.User
	result := database.DB.Preload("Wallet").First(&user, id)
	return &user, result.Error
}

func (*userRepo) UpdateWallet(user *entities.User) error {
	result := database.DB.Save(&user.Wallet)
	return result.Error
}

func (*userRepo) GetByUsername(username string) (*entities.User, error) {
	var user entities.User
	result := database.DB.Preload("Wallet").Where("username = ?", username).First(&user)
	return &user, result.Error
}
