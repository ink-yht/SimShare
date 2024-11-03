package service

import (
	"SimShare/internal/domain"
	"SimShare/internal/repository"
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicate             = repository.ErrDuplicate
	ErrInvalidUserOrPassword = errors.New("用户不存在或密码不对")
)

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

func (svc *UserService) Login(ctx context.Context, email string, password string) (domain.User, error) {
	user, err := svc.repo.FindByEmail(ctx, email)
	// err 两种情况
	// 1.系统错误
	// 2.用户没找到

	if err == repository.ErrRecordNotFound {
		return domain.User{}, ErrInvalidUserOrPassword
	}

	if err != nil {
		return domain.User{}, err
	}

	// 密码校验
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return user, nil
}

func (svc *UserService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	// 这个叫做快路径
	u, err := svc.repo.FindByPhone(ctx, phone)
	// 要判断有没有这个用户
	if err != repository.ErrRecordNotFound {
		// nil 会进来这里
		// 不为 ErrRecordNotFound 也会
		return u, err
	}

	// 在系统资源不足，触发降级之后，不执行慢路径了
	//if ctx.Value("降级") == "true" {
	//	return domain.User{}, errors.New("系统降级了")
	//}

	// 这个叫做慢路径
	// 你明确知道没有这个用户
	u = domain.User{
		Phone: phone,
	}
	err = svc.repo.Create(ctx, u)
	if err != nil && err != repository.ErrDuplicate {
		return u, err
	}
	// 这里会遇到主从延迟的问题
	return svc.repo.FindByPhone(ctx, phone)
}

func (svc *UserService) Profile(ctx context.Context, id int64) (domain.User, error) {
	u, err := svc.repo.FindById(ctx, id)
	return u, err
}
