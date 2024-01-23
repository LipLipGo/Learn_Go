package service

import (
	"Learn_Go/webook/internal/domain"
	"Learn_Go/webook/internal/repository"
	"context"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Signup(ctx context.Context, u domain.User) error
	Login(ctx context.Context, email string, password string) (domain.User, error)
	UpdateNonSensitiveInfo(ctx context.Context, u domain.User) error
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
	FindById(ctx context.Context, uid int64) (domain.User, error)
	FindOrCreateByWechat(ctx context.Context, info domain.WechatInfo) (domain.User, error)
}

type userService struct {
	repo repository.UserRepository
	//logger *zap.Logger
}

var (
	ErrDuplicateUser         = repository.ErrDuplicateEmail // 这里Email和Phone都是唯一索引，都会造成用户冲突，所以可以使用通用的错误名字
	ErrInvalidUserOrPassword = repository.ErrUserNotFound   // 账号或密码不正确，安全性更高
)

func NewuserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
		//logger: zap.L(),	// 这里可以传入自定义的 Logger ，但大部分情况，这里使用 zap.L() 就可以，
	}

}

func (svc *userService) Signup(ctx context.Context, u domain.User) error {

	// 密码加密
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.Create(ctx, u)
}

func (svc *userService) Login(ctx context.Context, email string, password string) (domain.User, error) {
	u, err := svc.repo.FindByEmail(ctx, email)

	if err == ErrInvalidUserOrPassword {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}

	// 检查密码是否正确

	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword // 当密码不正确时，也返回这个错误

	}
	return u, nil

}

func (svc *userService) UpdateNonSensitiveInfo(ctx context.Context, u domain.User) error {

	return svc.repo.UpdateNonZeroFields(ctx, u)
}

func (svc *userService) FindById(ctx context.Context, uid int64) (domain.User, error) {

	u, err := svc.repo.FindById(ctx, uid)
	if err != nil {
		return domain.User{}, err
	}

	return u, nil
}

func (svc *userService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {

	// 先找一下，我们认为，大部分用户是已经存在的用户
	u, err := svc.repo.FindByPhone(ctx, phone)

	// 这里直接判断这个错误是否是用户未找到，若不是，则有两种情况 1. nil，直接返回用户信息 2.系统错误
	if err != repository.ErrUserNotFound {
		return u, err
	}
	// 没有进去分支说明没找到用户，那么创建用户
	err = svc.repo.Create(ctx, domain.User{
		Phone: phone,
	})

	// 这里err有两种可能，一种是唯一索引冲突（Phone）
	// 一种是系统错误
	if err != nil && err != ErrDuplicateUser {
		return domain.User{}, err
	}

	// err要么nil，要么用户冲突
	// 直接查询用户
	// 但是注意，可能会有主从延迟，因为插入是在主库，查询是查从库，可能还没同步

	return svc.repo.FindByPhone(ctx, phone)
}

// 在service层进行密码加密（PBKDF2、BCrypt） ，同样的文本加密后的结果都不同

func (svc *userService) FindOrCreateByWechat(ctx context.Context, wechatInfo domain.WechatInfo) (domain.User, error) {

	// 这里因为我们开发的应用不存在多个应用，所以我们就直接使用OpenId
	u, err := svc.repo.FindByWechat(ctx, wechatInfo.OpenId)

	// 这里直接判断这个错误是否是用户未找到，若不是，则有两种情况 1. nil，直接返回用户信息 2.系统错误
	if err != repository.ErrUserNotFound {
		return u, err
	}
	// 没有进去分支说明没找到用户，那么创建用户
	zap.L().Info("新用户", zap.Any("wechatInfo", wechatInfo)) // 可以记录一下新用户
	err = svc.repo.Create(ctx, domain.User{
		WechatInfo: wechatInfo,
	})

	if err != nil && err != ErrDuplicateUser {
		return domain.User{}, err
	}

	return svc.repo.FindByWechat(ctx, wechatInfo.OpenId)
}
