package services

import (
	"banking-system/entities"
	"banking-system/models"
	"banking-system/repos"
	"fmt"

	log "github.com/sirupsen/logrus"
)

//go:generate mockgen -source=bankAccount.go -destination=mock/bankAccount.go

type BankAccountService interface {
	Create(req *models.CreateBankAccountRequest) (*models.BankAccountResponse, error)
	GetByID(id uint, userID uint) (*models.BankAccountResponse, error)
	GetByUserID(userID uint) ([]models.BankAccountResponse, error)
	Update(id uint, userID uint, req *models.UpdateBankAccountRequest) (*models.BankAccountResponse, error)
	Delete(id uint, userID uint) error
}

type bankAccountService struct {
	bankAccountRepo repos.BankAccountRepo
}

func NewBankAccountService(bankAccountRepo repos.BankAccountRepo) BankAccountService {
	return &bankAccountService{
		bankAccountRepo: bankAccountRepo,
	}
}

func (srv *bankAccountService) Create(req *models.CreateBankAccountRequest) (*models.BankAccountResponse, error) {
	bankAccount := &entities.BankAccount{
		UserID:        req.UserID,
		BankCode:      req.BankCode,
		AccountNumber: req.AccountNumber,
	}

	if err := srv.bankAccountRepo.Create(bankAccount); err != nil {
		log.Panicf("Failed to create bank account: %v", err)
	}

	return &models.BankAccountResponse{
		ID:            bankAccount.ID,
		UserID:        bankAccount.UserID,
		BankCode:      bankAccount.BankCode,
		AccountNumber: bankAccount.AccountNumber,
		CreatedAt:     bankAccount.CreatedAt,
		UpdatedAt:     bankAccount.UpdatedAt,
	}, nil
}

func (srv *bankAccountService) GetByID(id uint, userID uint) (*models.BankAccountResponse, error) {
	bankAccount, err := srv.bankAccountRepo.GetByID(id)
	if err != nil {
		if err.Error() == "record not found" {
			return nil, fmt.Errorf("bank account not found")
		}
		log.Panicf("Failed to get bank account: %v", err)
	}

	if bankAccount.UserID != userID {
		return nil, fmt.Errorf("unauthorized access to bank account")
	}

	return &models.BankAccountResponse{
		ID:            bankAccount.ID,
		UserID:        bankAccount.UserID,
		BankCode:      bankAccount.BankCode,
		AccountNumber: bankAccount.AccountNumber,
		CreatedAt:     bankAccount.CreatedAt,
		UpdatedAt:     bankAccount.UpdatedAt,
	}, nil
}

func (srv *bankAccountService) GetByUserID(userID uint) ([]models.BankAccountResponse, error) {
	bankAccounts, err := srv.bankAccountRepo.GetByUserID(userID)
	if err != nil {
		log.Panicf("Failed to get bank accounts: %v", err)
	}

	responses := make([]models.BankAccountResponse, len(bankAccounts))
	for i, ba := range bankAccounts {
		responses[i] = models.BankAccountResponse{
			ID:            ba.ID,
			UserID:        ba.UserID,
			BankCode:      ba.BankCode,
			AccountNumber: ba.AccountNumber,
			CreatedAt:     ba.CreatedAt,
			UpdatedAt:     ba.UpdatedAt,
		}
	}

	return responses, nil
}

func (srv *bankAccountService) Update(id uint, userID uint, req *models.UpdateBankAccountRequest) (*models.BankAccountResponse, error) {
	bankAccount, err := srv.bankAccountRepo.GetByID(id)
	if err != nil {
		if err.Error() == "record not found" {
			return nil, fmt.Errorf("bank account not found")
		}
		log.Panicf("Failed to get bank account: %v", err)
	}

	if bankAccount.UserID != userID {
		return nil, fmt.Errorf("unauthorized access to bank account")
	}

	bankAccount.BankCode = req.BankCode
	bankAccount.AccountNumber = req.AccountNumber

	if err := srv.bankAccountRepo.Update(bankAccount); err != nil {
		log.Panicf("Failed to update bank account: %v", err)
	}

	return &models.BankAccountResponse{
		ID:            bankAccount.ID,
		UserID:        bankAccount.UserID,
		BankCode:      bankAccount.BankCode,
		AccountNumber: bankAccount.AccountNumber,
		CreatedAt:     bankAccount.CreatedAt,
		UpdatedAt:     bankAccount.UpdatedAt,
	}, nil
}

func (srv *bankAccountService) Delete(id uint, userID uint) error {
	bankAccount, err := srv.bankAccountRepo.GetByID(id)
	if err != nil {
		if err.Error() == "record not found" {
			return fmt.Errorf("bank account not found")
		}
		log.Panicf("Failed to get bank account: %v", err)
	}

	if bankAccount.UserID != userID {
		return fmt.Errorf("unauthorized access to bank account")
	}

	if err := srv.bankAccountRepo.Delete(id); err != nil {
		log.Panicf("Failed to delete bank account: %v", err)
	}

	return nil
}
