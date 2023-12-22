package service

import (
	"Learn_Go/webook/internal/domain"
	"Learn_Go/webook/internal/repository"
	"context"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo *repository.UserRepository
}

var (
	ErrDuplicateEmail        = repository.ErrDuplicateEmail
	ErrInvalidUserOrPassword = repository.ErrUserNotFound // 账号或密码不正确，安全性更高
)

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}

}

func (svc *UserService) Signup(ctx context.Context, u domain.User) error {

	// 密码加密
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.Create(ctx, u)
}

func (svc *UserService) Login(ctx context.Context, email string, password string) (domain.User, error) {
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

func (svc *UserService) UpdateNonSensitiveInfo(ctx context.Context, u domain.User) error {
	return svc.repo.UpdateNonZeroFields(ctx, u)
}

// 在service层进行密码加密（PBKDF2、BCrypt） ，同样的文本加密后的结果都不同
