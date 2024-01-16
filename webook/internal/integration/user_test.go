package integration

import (
	"Learn_Go/webook/internal/integration/startup"
	"Learn_Go/webook/internal/web"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func init() {
	gin.SetMode(gin.ReleaseMode) // 防止gin的server输出过多日志，影响查看测试的日志
}

func TestUserHandler_SendSMSCode(t *testing.T) {
	rdb := startup.InitRedis()
	server := startup.InitWebServer()
	testCases := []struct {
		name   string
		before func(t *testing.T) // 准备数据
		after  func(t *testing.T) // 验证数据和清理数据
		phone  string

		wantCode int
		wantBody web.Result
	}{
		{
			name: "发送成功",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10) // 设置一个带过期时间的ctx
				defer cancel()
				key := "phone_code:login:15025635478"
				// 验证验证码是否存入了redis
				val, err := rdb.Get(ctx, key).Result()
				assert.NoError(t, err)
				// 因为我们不知道验证码到底是多少，因此通过val的长度来判断
				assert.True(t, len(val) > 0)
				// 验证验证码的过期时间是否正常
				dur, err := rdb.TTL(ctx, key).Result()
				assert.NoError(t, err)
				// 正常来说，验证码的过期时间是10分钟，但是执行代码还需要时间
				assert.True(t, dur > time.Minute*9+time.Second*50)
				// 验证后，将数据删除
				err = rdb.Del(ctx, key).Err()
				assert.NoError(t, err)
			},
			phone:    "15025635478",
			wantCode: http.StatusOK,
			wantBody: web.Result{
				Msg: "发送成功",
			},
		},
		{
			name:     "未输入手机号码",
			before:   func(t *testing.T) {},
			after:    func(t *testing.T) {},
			wantCode: http.StatusOK,
			wantBody: web.Result{
				Code: 4,
				Msg:  "请输入手机号！",
			},
		},
		{
			name: "发送太频繁",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10) // 设置一个带过期时间的ctx
				defer cancel()
				key := "phone_code:login:15025635478"
				// 提前存入一个验证码
				err := rdb.Set(ctx, key, "123456", time.Minute*10).Err()
				assert.NoError(t, err)

			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10) // 设置一个带过期时间的ctx
				defer cancel()
				key := "phone_code:login:15025635478"
				// 验证验证码是否存入了redis
				val, err := rdb.GetDel(ctx, key).Result()
				assert.NoError(t, err)
				assert.Equal(t, "123456", val)

			},
			phone:    "15025635478",
			wantCode: http.StatusOK,
			wantBody: web.Result{
				Code: 4,
				Msg:  "短信发送太频繁，请稍后再试",
			},
		},
		// 按理说，系统错误很难测试，因为系统错误一般是redis出问题，但是我们很难去模拟redis出问题，所以一般系统错误不测试
		{
			name: "系统错误",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10) // 设置一个带过期时间的ctx
				defer cancel()
				key := "phone_code:login:15025635478"
				// 提前存入一个验证码
				err := rdb.Set(ctx, key, "123456", 0).Err()
				assert.NoError(t, err)

			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10) // 设置一个带过期时间的ctx
				defer cancel()
				key := "phone_code:login:15025635478"
				// 验证验证码是否存入了redis
				val, err := rdb.GetDel(ctx, key).Result()
				assert.NoError(t, err)
				assert.Equal(t, "123456", val)

			},
			phone:    "15025635478",
			wantCode: http.StatusOK,
			wantBody: web.Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			tc.before(t)
			defer tc.after(t)

			req, err := http.NewRequest(http.MethodPost,
				"/users/login_sms/code/send",
				bytes.NewReader([]byte(fmt.Sprintf(`{"phone":"%s"}`, tc.phone))))
			req.Header.Set("Content-Type", "application/json")
			assert.NoError(t, err)

			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, req)

			assert.Equal(t, tc.wantCode, recorder.Code)
			if tc.wantCode != http.StatusOK {
				return
			}
			// 反序列化为结构体
			var res web.Result
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantBody, res)

		})
	}
}
