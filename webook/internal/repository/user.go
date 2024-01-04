package repository

import (
	"Learn_Go/webook/internal/domain"
	"Learn_Go/webook/internal/repository/dao"
	"context"
	"github.com/gin-gonic/gin"
	"time"
)

var (
	ErrDuplicateEmail = dao.ErrDuplicateEmail
	ErrUserNotFound   = dao.ErrRecordNotFound
)

type UserRepository struct {
	dao *dao.UserDao
}

func NewUserRepository(dao *dao.UserDao) *UserRepository {
	return &UserRepository{
		dao: dao,
	}
}

func (repo *UserRepository) Create(ctx context.Context, u domain.User) error {
	return repo.dao.Insert(ctx, dao.User{
		Email:    u.Email,
		Password: u.Password,
	})
}

func (repo *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := repo.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return repo.toDomain(u), nil

}

func (repo *UserRepository) toDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
		AboutMe:  u.AboutMe,
		BirthDay: time.UnixMilli(u.Birthday),
		NickName: u.Nickname,
	}
}

func (repo *UserRepository) UpdateNonZeroFields(ctx context.Context, u domain.User) error {
	return repo.dao.UpdateById(ctx, repo.toEntity(u))
}

func (repo *UserRepository) toEntity(u domain.User) dao.User {
	return dao.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
		Birthday: u.BirthDay.UnixMilli(),
		AboutMe:  u.AboutMe,
		Nickname: u.NickName,
	}
}

func (repo *UserRepository) FindById(ctx *gin.Context, uid int64) (domain.User, error) {
	u, err := repo.dao.FindById(ctx, uid)
	if err != nil {
		return domain.User{}, err
	}

	return repo.toDomain(u), nil

}
