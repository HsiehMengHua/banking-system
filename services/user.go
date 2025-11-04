package services

import (
	"banking-system/entities"
	"banking-system/models"
	"banking-system/repos"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

//go:generate mockgen -source=user.go -destination=mock/user.go

type UserService interface {
	Register(req *models.RegisterRequest) error
	GetByUsername(username string) (*entities.User, error)
	GetByID(id uint) (*models.UserInfoResponse, error)
}

type userService struct {
	userRepo repos.UserRepo
}

func NewUserService(userRepo repos.UserRepo) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (srv *userService) Register(req *models.RegisterRequest) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Panicf("Failed to hash password: %v", err)
	}

	user := &entities.User{
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
		Name:         req.Name,
		Wallet: entities.Wallet{
			Currency: "TWD",
			Balance:  0,
		},
	}

	if err := srv.userRepo.Create(user); err != nil {
		if strings.Contains(err.Error(), "uni_users_username") || strings.Contains(err.Error(), "duplicate key") {
			return fmt.Errorf("username '%s' is already taken", req.Username)
		}
		log.Panicf("Failed to create user: %v", err)
	}

	return nil
}

func (srv *userService) GetByUsername(username string) (*entities.User, error) {
	userEntity, err := srv.userRepo.GetByUsername(username)

	if err != nil {
		if err.Error() == "record not found" {
			return nil, err
		} else {
			log.Panicf("Failed to get user by username: %v", err)
		}
	}
	return userEntity, nil
}

func (srv *userService) GetByID(id uint) (*models.UserInfoResponse, error) {
	user, err := srv.userRepo.Get(id)
	if err != nil {
		if err.Error() == "record not found" {
			return nil, fmt.Errorf("user not found")
		}
		log.Panicf("Failed to get user by ID: %v", err)
	}

	return &models.UserInfoResponse{
		Username: user.Username,
		Balance:  user.Wallet.Balance,
	}, nil
}
