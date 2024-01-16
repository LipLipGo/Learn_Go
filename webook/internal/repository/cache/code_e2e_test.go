package cache

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRedisCodeCache_Set_e2e(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		ctx   context.Context
		biz   string
		phone string
		code  string

		wantErr error
	}{
		{
			name: "设置成功",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10) // 设置一个带过期时间的ctx
				defer cancel()
				key := "phone_code:test:15023154562"

				// 验证验证码的过期时间是否正常
				dur, err := rdb.TTL(ctx, key).Result()
				assert.NoError(t, err)
				// 正常来说，验证码的过期时间是10分钟，但是执行代码还需要时间
				assert.True(t, dur > time.Minute*9+time.Second*50)
				// 验证验证码是否存入了redis
				val, err := rdb.Get(ctx, key).Result()
				assert.NoError(t, err)
				// 判断验证码是否正确
				assert.Equal(t, "123456", val)
				// 验证后，将数据删除
				err = rdb.Del(ctx, key).Err()
				assert.NoError(t, err)
			},
			ctx:     context.Background(),
			biz:     "test",
			phone:   "15023154562",
			code:    "123456",
			wantErr: nil,
		},
		{
			name: "发送太频繁",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10) // 设置一个带过期时间的ctx
				defer cancel()
				key := "phone_code:test:15023154562"
				// 提前存入一个验证码
				err := rdb.Set(ctx, key, "654321", time.Minute*10).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10) // 设置一个带过期时间的ctx
				defer cancel()
				key := "phone_code:test:15023154562"

				// 验证验证码的过期时间是否正常
				dur, err := rdb.TTL(ctx, key).Result()
				assert.NoError(t, err)
				// 正常来说，验证码的过期时间是10分钟，但是执行代码还需要时间
				assert.True(t, dur > time.Minute*9+time.Second*50)
				// 验证验证码是否存入了redis
				val, err := rdb.GetDel(ctx, key).Result()
				assert.NoError(t, err)
				// 判断验证码是否正确
				assert.Equal(t, "654321", val)

			},
			ctx:     context.Background(),
			biz:     "test",
			phone:   "15023154562",
			code:    "123456",
			wantErr: ErrCodeSendTooMany,
		},
		{
			name: "系统错误",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10) // 设置一个带过期时间的ctx
				defer cancel()
				key := "phone_code:test:15023154562"
				// 提前存入一个验证码
				err := rdb.Set(ctx, key, "654321", 0).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10) // 设置一个带过期时间的ctx
				defer cancel()
				key := "phone_code:test:15023154562"
				// 验证验证码是否存入了redis
				val, err := rdb.GetDel(ctx, key).Result()
				assert.NoError(t, err)
				// 判断验证码是否正确
				assert.Equal(t, "654321", val)

			},
			ctx:     context.Background(),
			biz:     "test",
			phone:   "15023154562",
			code:    "123456",
			wantErr: errors.New("验证码存在，但没有过期时间！"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)
			c := NewRedisCodeCache(rdb)
			err := c.Set(tc.ctx, tc.biz, tc.phone, tc.code)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
