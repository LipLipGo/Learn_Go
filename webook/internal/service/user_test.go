package service

import (
	"Learn_Go/webook/internal/domain"
	"Learn_Go/webook/internal/repository"
	repomocks "Learn_Go/webook/internal/repository/mocks"
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func TestEncrypt(t *testing.T) {
	password := []byte("123456&lip")
	encrypted, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	assert.NoError(t, err)

	fmt.Println(string(encrypted))
	err = bcrypt.CompareHashAndPassword(encrypted, []byte("123456&lip"))
	assert.NoError(t, err)
}

func Test_userService_Login(t *testing.T) {
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) repository.UserRepository

		// 预期输入
		ctx      context.Context
		Email    string
		Password string

		// 预期返回值
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "登陆成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "12345@qq.com").Return(domain.User{
					Email: "12345@qq.com",
					// 这里拿到的密码应该是加密后的密码
					Password: "$2a$10$Xtf2o6ErMJcGNsdVcAJln.5qcQN4GzHOX4DIhPAOzHB.DF3lEzaVu",
					Phone:    "15023113254",
				}, nil)
				return repo
			},
			Email: "12345@qq.com",
			// 用户输入的，未加密的
			Password: "123456&lip",

			wantUser: domain.User{
				Email:    "12345@qq.com",
				Password: "$2a$10$Xtf2o6ErMJcGNsdVcAJln.5qcQN4GzHOX4DIhPAOzHB.DF3lEzaVu",
				Phone:    "15023113254",
			},
			wantErr: nil,
		},
		{
			name: "用户未注册",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "12345@qq.com").Return(domain.User{}, ErrInvalidUserOrPassword)
				return repo
			},
			Email:    "12345@qq.com",
			Password: "123456&lip",

			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "12345@qq.com").Return(domain.User{}, errors.New("db错误"))
				return repo
			},
			Email: "12345@qq.com",
			// 用户输入的，未加密的
			Password: "123456&lip",

			wantUser: domain.User{},
			wantErr:  errors.New("db错误"),
		},
		{
			name: "密码错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "12345@qq.com").Return(domain.User{
					Email: "12345@qq.com",
					// 这里拿到的密码应该是加密后的密码
					Password: "$2a$10$Xtf2o6ErMJcGNsdVcAJln.5qcQN4GzHOX4DIhPAOzHB.DF3lEzaVu",
					Phone:    "15023113254",
				}, nil)
				return repo
			},
			Email: "12345@qq.com",
			// 用户输入的，未加密的
			Password: "123456&li",

			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := tc.mock(ctrl)
			svc := NewuserService(repo)
			User, err := svc.Login(tc.ctx, tc.Email, tc.Password)
			assert.Equal(t, tc.wantUser, User)
			assert.Equal(t, tc.wantErr, err)

		})
	}
}
