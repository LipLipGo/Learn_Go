package cache

import (
	"Learn_Go/webook/internal/repository/cache/redismocks"
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestRedisCodeCache_Set(t *testing.T) {
	keyFunc := func(biz, phone string) string {
		return fmt.Sprintf("phone_code:%s:%s", biz, phone)
	}
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) redis.Cmdable

		ctx     context.Context
		biz     string
		phone   string
		code    string
		wantErr error
	}{
		{
			name: "设置成功",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				res := redismocks.NewMockCmdable(ctrl)
				cmd := redis.NewCmd(context.Background())
				cmd.SetVal(int64(0)) // 这里只能使用int64
				cmd.SetErr(nil)
				// 这里其实返回的是cmd，通过cmd设置value和err
				res.EXPECT().Eval(gomock.Any(), luaSetCode, []string{keyFunc("test", "15023154562")}, []any{"123456"}).Return(cmd)
				return res

			},
			ctx:     context.Background(),
			biz:     "test",
			phone:   "15023154562",
			code:    "123456",
			wantErr: nil,
		},
		{
			name: "redis返回err",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				res := redismocks.NewMockCmdable(ctrl)
				cmd := redis.NewCmd(context.Background())
				cmd.SetVal(int64(0)) // 这里只能使用int64
				cmd.SetErr(errors.New("redis错误"))
				// 这里其实返回的是cmd，通过cmd设置value和err
				res.EXPECT().Eval(gomock.Any(), luaSetCode, []string{keyFunc("test", "15023154562")}, []any{"123456"}).Return(cmd)
				return res

			},
			ctx:     context.Background(),
			biz:     "test",
			phone:   "15023154562",
			code:    "123456",
			wantErr: errors.New("redis错误"),
		},
		{
			name: "验证码存在，没有过期时间",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				res := redismocks.NewMockCmdable(ctrl)
				cmd := redis.NewCmd(context.Background())
				cmd.SetVal(int64(-2)) // 这里只能使用int64
				cmd.SetErr(nil)
				// 这里其实返回的是cmd，通过cmd设置value和err
				res.EXPECT().Eval(gomock.Any(), luaSetCode, []string{keyFunc("test", "15023154562")}, []any{"123456"}).Return(cmd)
				return res

			},
			ctx:     context.Background(),
			biz:     "test",
			phone:   "15023154562",
			code:    "123456",
			wantErr: errors.New("验证码存在，但没有过期时间！"),
		},
		{
			name: "验证码发送太频繁",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				res := redismocks.NewMockCmdable(ctrl)
				cmd := redis.NewCmd(context.Background())
				cmd.SetVal(int64(-1)) // 这里只能使用int64
				cmd.SetErr(nil)
				// 这里其实返回的是cmd，通过cmd设置value和err
				// 这里调用EVAL需要注意传入参数的格式，最后一个接收的是接口的不定参数，所以这个应该是any的不定参数
				res.EXPECT().Eval(gomock.Any(), luaSetCode, []string{keyFunc("test", "15023154562")}, []any{"123456"}).Return(cmd)
				return res

			},
			ctx:     context.Background(),
			biz:     "test",
			phone:   "15023154562",
			code:    "123456",
			wantErr: ErrCodeSendTooMany,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cmd := tc.mock(ctrl)
			cc := NewRedisCodeCache(cmd)
			err := cc.Set(tc.ctx, tc.biz, tc.phone, tc.code)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
