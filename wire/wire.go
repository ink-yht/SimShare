//go:build wireinject

// 让 wire 来注入这里的代码
package wire

import (
	"SimShare/wire/repository"
	"SimShare/wire/repository/dao"
	"github.com/google/wire"
)

func InitRepository() *repository.UserRepository {
	// 这个方法里面传入各个组件的初始化方法
	wire.Build(repository.NewUserRepository, dao.NewUserDao, InitDB)
	return new(repository.UserRepository)
}
