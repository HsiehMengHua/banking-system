package repos

import (
	"banking-system/database"
	"banking-system/entities"
)

//go:generate mockgen -source=userRepo.go -destination=mock/userRepo.go

type UserRepo interface {
	Get(id uint) (*entities.User, error)
	UpdateWallet(user *entities.User) error
}

type userRepo struct {
}

func NewUserRepo() UserRepo {
	return &userRepo{}
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
