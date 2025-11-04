package repos

import (
	"banking-system/database"
	"banking-system/entities"
)

//go:generate mockgen -source=bankAccountRepo.go -destination=mock/bankAccountRepo.go

type BankAccountRepo interface {
	Create(bankAccount *entities.BankAccount) error
	GetByID(id uint) (*entities.BankAccount, error)
	GetByUserID(userID uint) ([]entities.BankAccount, error)
	Update(bankAccount *entities.BankAccount) error
	Delete(id uint) error
}

type bankAccountRepo struct {
}

func NewBankAccountRepo() BankAccountRepo {
	return &bankAccountRepo{}
}

func (*bankAccountRepo) Create(bankAccount *entities.BankAccount) error {
	result := database.DB.Create(bankAccount)
	return result.Error
}

func (*bankAccountRepo) GetByID(id uint) (*entities.BankAccount, error) {
	var bankAccount entities.BankAccount
	result := database.DB.First(&bankAccount, id)
	return &bankAccount, result.Error
}

func (*bankAccountRepo) GetByUserID(userID uint) ([]entities.BankAccount, error) {
	var bankAccounts []entities.BankAccount
	result := database.DB.Where("user_id = ?", userID).Find(&bankAccounts)
	return bankAccounts, result.Error
}

func (*bankAccountRepo) Update(bankAccount *entities.BankAccount) error {
	result := database.DB.Save(bankAccount)
	return result.Error
}

func (*bankAccountRepo) Delete(id uint) error {
	result := database.DB.Delete(&entities.BankAccount{}, id)
	return result.Error
}
