package service

import (
	"SimShare/internal/domain"
	"SimShare/internal/repository"
	"context"
	"golang.org/x/crypto/bcrypt"
)

var ErrDuplicateEmail = repository.ErrDuplicateEmail

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (svc *UserService) SignUp(ctx context.Context, user domain.User) error {
	// 加密，然后存起来
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hash)

	// 存起来
	return svc.repo.Create(ctx, user)
}
