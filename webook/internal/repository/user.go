package repository

import (
	"Learn_Go/webook/internal/domain"
	"Learn_Go/webook/internal/repository/cache"
	"Learn_Go/webook/internal/repository/dao"
	"context"
	"database/sql"
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

var (
	ErrDuplicateEmail = dao.ErrDuplicateEmail
	ErrUserNotFound   = dao.ErrRecordNotFound
)

type UserRepository interface {
	Create(ctx context.Context, u domain.User) error
	UpdateNonZeroFields(ctx context.Context, u domain.User) error
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	FindById(ctx context.Context, uid int64) (domain.User, error)
	FindByEmail(ctx context.Context, email string) (domain.User, error)
}

type CachedUserRepository struct {
	dao   dao.UserDao
	cache cache.UserCache
}

func NewCachedUserRepository(dao dao.UserDao, c cache.UserCache) UserRepository {
	return &CachedUserRepository{
		dao:   dao,
		cache: c,
	}
}

func (repo *CachedUserRepository) Create(ctx context.Context, u domain.User) error {
	return repo.dao.Insert(ctx, repo.toEntity(u))
}

func (repo *CachedUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := repo.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return repo.toDomain(u), nil

}

func (repo *CachedUserRepository) toDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email.String,
		Phone:    u.Phone.String,
		Password: u.Password,
		AboutMe:  u.AboutMe,
		BirthDay: time.UnixMilli(u.Birthday),
		NickName: u.Nickname,
		Ctime:    time.UnixMilli(u.Ctime),
	}
}

func (repo *CachedUserRepository) UpdateNonZeroFields(ctx context.Context, u domain.User) error {
	return repo.dao.UpdateById(ctx, repo.toEntity(u))
}

func (repo *CachedUserRepository) toEntity(u domain.User) dao.User {
	return dao.User{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			Valid:  u.Email != "",
		},
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		Password: u.Password,
		Birthday: u.BirthDay.UnixMilli(),
		AboutMe:  u.AboutMe,
		Nickname: u.NickName,
	}
}

func (repo *CachedUserRepository) FindById(ctx context.Context, uid int64) (domain.User, error) {
	du, err := repo.cache.Get(ctx, uid)

	// 只要err为nil，就返回
	if err == nil {
		return du, err
	}
	// err不为nil，就要查询数据库
	// err有两种可能
	// 1. key不存在，说明 redis 是正常的，uid可能正确也可能不正确
	// 2. 访问 redis 有问题。可能是网路有问题，也可能是 redis 本身就崩溃了

	u, err := repo.dao.FindById(ctx, uid)
	if err != nil {
		return domain.User{}, err
	}
	du = repo.toDomain(u)

	err = repo.cache.Set(ctx, du)
	if err != nil {
		// redis 有问题。可能是网路有问题，也可能是 redis 本身就崩溃了。如果这里出问题了，那么下次查询还是会查数据库，这种现象叫缓存击穿，那么数据库的压力也会很大
		// 如果缓存写入失败，我们不中断，输出一下错误就好
		//return domain.User{}, err
		log.Println(err)
	}

	return du, nil

}

// 查询的另一种写法，进一步判定err是何种错误

func (repo *CachedUserRepository) FindByIdV1(ctx *gin.Context, uid int64) (domain.User, error) {
	du, err := repo.cache.Get(ctx, uid)

	// 只要err为nil，就返回
	if err == nil {
		return du, err
	}
	// 进一步判断是何种错误，只有当key不存在时才会去查询数据库，否则直接返回错误
	switch err {
	case nil:
		return du, nil
	case cache.ErrKeyNotExist: // 缓存没有数据，但是redis是正常的
		u, err := repo.dao.FindById(ctx, uid)
		if err != nil {
			return domain.User{}, err
		}
		du = repo.toDomain(u)

		err = repo.cache.Set(ctx, du)
		if err != nil {
			// redis 有问题。可能是网路有问题，也可能是 redis 本身就崩溃了。如果这里出问题了，那么下次查询还是会查数据库，这种现象叫缓存击穿，那么数据库的压力也会很大
			return domain.User{}, err
		}
		return du, nil
	default: // 缓存有数据，但是redis是不正常的，降级写法，redis不正常，不去查数据库，因为数据库有很多业务，不要把数据库打爆
		return domain.User{}, err

	}

}

func (repo *CachedUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := repo.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return repo.toDomain(u), nil

}
