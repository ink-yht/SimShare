package repository

import (
	"SimShare/internal/domain"
	"SimShare/internal/repository/cache"
	"SimShare/internal/repository/dao"
	"context"
	"database/sql"
	"time"
)

var (
	ErrDuplicate      = dao.ErrDuplicate
	ErrRecordNotFound = dao.ErrRecordNotFound
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	FindById(ctx context.Context, id int64) (domain.User, error)
}

type CacheUserRepository struct {
	dao   dao.UserDao
	cache cache.UserCache
}

func NewUserRepository(dao dao.UserDao, cache cache.UserCache) UserRepository {
	return &CacheUserRepository{
		dao:   dao,
		cache: cache,
	}
}

func (repo *CacheUserRepository) Create(ctx context.Context, user domain.User) error {
	return repo.dao.Insert(ctx, repo.domainToEntity(user))
}

func (repo *CacheUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := repo.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return repo.entityToDomain(user), err
}

func (repo *CacheUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	user, err := repo.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return repo.entityToDomain(user), err
}

//func (repo *CacheUserRepository) toDomain(u dao.User) domain.User {
//	return domain.User{
//		Id:       u.Id,
//		Email:    u.Email,
//		Password: u.Password,
//	}
//}

func (repo *CacheUserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
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

	u = repo.entityToDomain(ue)

	go func() {
		err = repo.cache.Set(ctx, u)
		if err != nil {
			// 打日志，做监控
		}
	}()

	return u, err
}

func (repo *CacheUserRepository) domainToEntity(u domain.User) dao.User {
	return dao.User{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			// 确实有手机号
			Valid: u.Email != "",
		},
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		Password: u.Password,
		Ctime:    u.Ctime.UnixMilli(),
	}
}

func (repo *CacheUserRepository) entityToDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email.String,
		Password: u.Password,
		Phone:    u.Phone.String,
		Ctime:    time.UnixMilli(u.Ctime),
	}
}
