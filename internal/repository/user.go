package repository

import (
	"SimShare/internal/domain"
	"SimShare/internal/repository/cache"
	"SimShare/internal/repository/dao"
	"context"
)

var (
	ErrDuplicateEmail = dao.ErrDuplicateEmail
	ErrRecordNotFound = dao.ErrRecordNotFound
)

type UserRepository struct {
	dao   *dao.UserDAO
	cache *cache.UserCache
}

func NewUserRepository(dao *dao.UserDAO, cache *cache.UserCache) *UserRepository {
	return &UserRepository{
		dao:   dao,
		cache: cache,
	}
}

func (repo *UserRepository) Create(ctx context.Context, user domain.User) error {
	return repo.dao.Insert(ctx, dao.User{
		Email:    user.Email,
		Password: user.Password,
	})
}

func (repo *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := repo.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return repo.toDomain(user), err
}

func (repo *UserRepository) toDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
	}
}

func (repo *UserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	u, err := repo.cache.Get(ctx, id)
	// 缓存里面有数据
	// 缓存里面没有数据
	// 缓存出错了，不知道有没有数据
	if err == nil {
		// 必然有数据
		return u, err
	}

	// 没这个数据
	if err == cache.ErrKeyNotExist {
		// 去数据库里面加载
	}

	/*
	   这里怎么办? err = io.EOF
	   要不要去数据库加载

	   选加载 - 做好兜底，万一 redis 崩了，要保护住数据库
	   数据库限流

	   不加载 - 用户体验差
	*/

	ue, err := repo.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	u = domain.User{
		Id:       ue.Id,
		Email:    ue.Email,
		Password: ue.Password,
	}

	go func() {
		err = repo.cache.Set(ctx, u)
		if err != nil {
			// 打日志，做监控
		}
	}()

	return u, err
}
