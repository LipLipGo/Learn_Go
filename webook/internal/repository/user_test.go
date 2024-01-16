package repository

import (
	"Learn_Go/webook/internal/domain"
	"Learn_Go/webook/internal/repository/cache"
	cachemocks "Learn_Go/webook/internal/repository/cache/mocks"
	"Learn_Go/webook/internal/repository/dao"
	daomocks "Learn_Go/webook/internal/repository/dao/mocks"
	"context"
	"database/sql"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestCachedUserRepository_FindById(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) (dao.UserDao, cache.UserCache)

		ctx      context.Context
		uid      int64
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "查找成功，缓存未命中",
			mock: func(ctrl *gomock.Controller) (dao.UserDao, cache.UserCache) {
				uid := int64(123)
				d := daomocks.NewMockUserDao(ctrl)
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), uid).Return(domain.User{}, cache.ErrKeyNotExist)
				d.EXPECT().FindById(gomock.Any(), uid).Return(dao.User{
					Id:       uid,
					Email:    sql.NullString{String: "123", Valid: true},
					Password: "123",
					Ctime:    101,
					Utime:    102,
					Birthday: 100,
					AboutMe:  "自我介绍",
					Nickname: "lip",
					Phone:    sql.NullString{String: "15012354652", Valid: true},
				}, nil)
				c.EXPECT().Set(gomock.Any(), domain.User{
					Id:       123,
					Email:    "123",
					Phone:    "15012354652",
					Password: "123",
					AboutMe:  "自我介绍",
					BirthDay: time.UnixMilli(100),
					NickName: "lip",
					Ctime:    time.UnixMilli(101),
				}).Return(nil)
				return d, c
			},

			ctx: context.Background(),
			uid: 123,
			wantUser: domain.User{
				Id:       123,
				Email:    "123",
				Phone:    "15012354652",
				Password: "123",
				AboutMe:  "自我介绍",
				BirthDay: time.UnixMilli(100),
				NickName: "lip",
				Ctime:    time.UnixMilli(101),
			},
			wantErr: nil,
		},
		{
			name: "缓存命中",
			mock: func(ctrl *gomock.Controller) (dao.UserDao, cache.UserCache) {
				uid := int64(123)
				d := daomocks.NewMockUserDao(ctrl)
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), uid).Return(domain.User{
					Id:       123,
					Email:    "123",
					Phone:    "15012354652",
					Password: "123",
					AboutMe:  "自我介绍",
					BirthDay: time.UnixMilli(100),
					NickName: "lip",
					Ctime:    time.UnixMilli(101),
				}, nil)

				return d, c
			},

			ctx: context.Background(),
			uid: 123,
			wantUser: domain.User{
				Id:       123,
				Email:    "123",
				Phone:    "15012354652",
				Password: "123",
				AboutMe:  "自我介绍",
				BirthDay: time.UnixMilli(100),
				NickName: "lip",
				Ctime:    time.UnixMilli(101),
			},
			wantErr: nil,
		},
		{
			name: "未找到用户",
			mock: func(ctrl *gomock.Controller) (dao.UserDao, cache.UserCache) {
				uid := int64(123)
				d := daomocks.NewMockUserDao(ctrl)
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), uid).Return(domain.User{}, cache.ErrKeyNotExist)
				d.EXPECT().FindById(gomock.Any(), uid).Return(dao.User{}, dao.ErrRecordNotFound)
				return d, c
			},

			ctx:      context.Background(),
			uid:      123,
			wantUser: domain.User{},
			wantErr:  ErrUserNotFound,
		},
		{
			name: "回写缓存错误",
			mock: func(ctrl *gomock.Controller) (dao.UserDao, cache.UserCache) {
				uid := int64(123)
				d := daomocks.NewMockUserDao(ctrl)
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), uid).Return(domain.User{}, cache.ErrKeyNotExist)
				d.EXPECT().FindById(gomock.Any(), uid).Return(dao.User{
					Id:       uid,
					Email:    sql.NullString{String: "123", Valid: true},
					Password: "123",
					Ctime:    101,
					Utime:    102,
					Birthday: 100,
					AboutMe:  "自我介绍",
					Nickname: "lip",
					Phone:    sql.NullString{String: "15012354652", Valid: true},
				}, nil)
				c.EXPECT().Set(gomock.Any(), domain.User{
					Id:       123,
					Email:    "123",
					Phone:    "15012354652",
					Password: "123",
					AboutMe:  "自我介绍",
					BirthDay: time.UnixMilli(100),
					NickName: "lip",
					Ctime:    time.UnixMilli(101),
				}).Return(errors.New("redis错误"))
				return d, c
			},

			ctx: context.Background(),
			uid: 123,
			wantUser: domain.User{
				Id:       123,
				Email:    "123",
				Phone:    "15012354652",
				Password: "123",
				AboutMe:  "自我介绍",
				BirthDay: time.UnixMilli(100),
				NickName: "lip",
				Ctime:    time.UnixMilli(101),
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			userDao, userCache := tc.mock(ctrl)
			repo := NewCachedUserRepository(userDao, userCache)
			User, err := repo.FindById(tc.ctx, tc.uid)
			assert.Equal(t, tc.wantUser, User)
			assert.Equal(t, tc.wantErr, err)

		})
	}

}
