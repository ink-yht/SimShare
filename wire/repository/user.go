package repository

import "SimShare/wire/repository/dao"

type UserRepository struct {
	dao *dao.UserDao
}

func NewUserRepository(dao *dao.UserDao) *UserRepository {
	return &UserRepository{
		dao: dao,
	}

}
