//go:build wireinject

package wire

import (
	"Learn_Go/wire/repository"
	"Learn_Go/wire/repository/dao"
	"github.com/google/wire"
)

func InitUserRepository() *repository.UserRepository {
	wire.Build(repository.NewUserRepository, dao.NewUserDAO, dao.InitDB) // 通过wire生成初始化代码
	return &repository.UserRepository{}
}
