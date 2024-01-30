package web

import (
	"Learn_Go/webook/internal/domain"
	"Learn_Go/webook/internal/repository"
	"Learn_Go/webook/internal/service"
	svcmocks "Learn_Go/webook/internal/service/mocks"
	ijwt "Learn_Go/webook/internal/web/jwt"
	"bytes"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

//func TestHTTP(t *testing.T) {
//	// 构造HTTP请求，如果没有请求体就传一个 nil
//	req, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewReader([]byte("我的请求体")))
//	assert.NoError(t, err)
//	recorder := httptest.NewRecorder()
//	assert.Equal(t, http.StatusOK, recorder.Code) // 断言返回的响应码是否是200
//	h := NewUserHandler(nil, nil)
//}

//func TestUserHandler_SignUp(t *testing.T) {
//	testCases := []struct {
//		name       string
//		mock       func(ctrl *gomock.Controller) (service.UserService, service.CodeService)
//		reqBuilder func(t *testing.T) *http.Request
//		wangCode   int
//		wangBody   string
//	}{
//		{
//			name: "注册成功",
//			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
//				userSvc := svcmocks.NewMockUserService(ctrl)
//				userSvc.EXPECT().Signup(gomock.Any(), domain.User{
//					Email:    "12345@qq.com",
//					Password: "123456&lip",
//				}).Return(nil)
//
//				codeSvc := svcmocks.NewMockCodeService(ctrl)
//
//				return userSvc, codeSvc
//			},
//
//			reqBuilder: func(t *testing.T) *http.Request {
//				req, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewReader([]byte(`{
//"email":"12345@qq.com",
//"password":"123456&lip",
//"confirmPassword":"123456&lip"
//}`)))
//				assert.NoError(t, err)
//				req.Header.Set("Content-Type", "application/json")
//				return req
//			},
//
//			wangCode: http.StatusOK,
//			wangBody: "注册成功！",
//		},
//	}
//
//	for _, tc := range testCases {
//		t.Run(tc.name, func(t *testing.T) {
//			ctrl := gomock.NewController(t)
//			defer ctrl.Finish()
//
//			userSvc, codeSvc := tc.mock(ctrl)
//
//			server := gin.Default()
//			hdl := NewUserHandler(userSvc, codeSvc)
//			hdl.RegisterRouters(server)
//
//			req := tc.reqBuilder
//
//			recorder := httptest.NewRecorder()
//			server.ServeHTTP(recorder, req(t))
//
//			assert.Equal(t, tc.wangCode, recorder.Code)
//			assert.Equal(t, tc.wangBody, recorder.Body)
//
//		})
//	}
//
//}

func TestUserHandler_SignUp(t *testing.T) {
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) (service.UserService, service.CodeService)

		reqBuilder func(t *testing.T) *http.Request

		wantCode int
		wantBody string
	}{
		// 注册成功测试用例
		{
			name: "注册成功",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().Signup(gomock.Any(), domain.User{
					Email:    "1234@qq.com",
					Password: "123456&lip",
				}).Return(nil)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewReader([]byte(`{
"email":"1234@qq.com",
"password":"123456&lip",
"confirmPassword":"123456&lip"
}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},

			wantCode: http.StatusOK,
			wantBody: "注册成功！",
		},

		// 异常流程测试用例
		{
			name: "Bind出错",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				// Bind出错，不会进入这个分支
				//userSvc.EXPECT().Signup(gomock.Any(), domain.User{
				//	Email:    "1234@qq.com",
				//	Password: "123456&lip",
				//}).Return(nil)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewReader([]byte(`{
"email":"1234@qq.com",
"password":"123456
}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},

			wantCode: http.StatusBadRequest,
		},
		{
			name: "两次密码输入不一致",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				// 邮箱格式不对，不会进入这个分支
				//userSvc.EXPECT().Signup(gomock.Any(), domain.User{
				//	Email:    "1234@qq.com",
				//	Password: "123456&lip",
				//}).Return(nil)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewReader([]byte(`{
"email":"1234",
"password":"123456&lip",
"confirmPassword":"123456&lip"
}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},

			wantCode: http.StatusOK,
			wantBody: "非法邮箱格式！",
		},
		{
			name: "邮箱格式不对",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				// 邮箱格式不对，不会进入这个分支
				//userSvc.EXPECT().Signup(gomock.Any(), domain.User{
				//	Email:    "1234@qq.com",
				//	Password: "123456&lip",
				//}).Return(nil)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewReader([]byte(`{
"email":"12345@qq.com",
"password":"123456&lip",
"confirmPassword":"123456&li"
}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},

			wantCode: http.StatusOK,
			wantBody: "两次密码输入不一致！",
		},
		{
			name: "密码格式不对",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				// 邮箱格式不对，不会进入这个分支
				//userSvc.EXPECT().Signup(gomock.Any(), domain.User{
				//	Email:    "1234@qq.com",
				//	Password: "123456&lip",
				//}).Return(nil)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewReader([]byte(`{
"email":"12345@qq.com",
"password":"12345",
"confirmPassword":"12345"
}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},

			wantCode: http.StatusOK,
			wantBody: "密码必须包含字母、数字、特殊字符，并且不少于八位",
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().Signup(gomock.Any(), domain.User{
					Email:    "1234@qq.com",
					Password: "123456&lip",
				}).Return(errors.New("db错误"))
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewReader([]byte(`{
"email":"1234@qq.com",
"password":"123456&lip",
"confirmPassword":"123456&lip"
}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},

			wantCode: http.StatusOK,
			wantBody: "系统错误",
		},
		{
			name: "邮箱冲突",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().Signup(gomock.Any(), domain.User{
					Email:    "1234@qq.com",
					Password: "123456&lip",
				}).Return(repository.ErrDuplicateEmail)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewReader([]byte(`{
"email":"1234@qq.com",
"password":"123456&lip",
"confirmPassword":"123456&lip"
}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},

			wantCode: http.StatusOK,
			wantBody: "该邮箱已被注册，请更换一个邮箱！",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userSvc, codeSvc := tc.mock(ctrl)

			server := gin.Default()
			h := NewUserHandler(userSvc, codeSvc, ijwt.NewRedisJWTHandler(redis.NewClient(&redis.Options{Addr: ""})))
			h.RegisterRouters(server)

			req := tc.reqBuilder(t)

			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, req)

			assert.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantBody, recorder.Body.String())

		})
	}

}

func TestMock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	// mock实现，模拟实现
	userSvc := svcmocks.NewMockUserService(ctrl)
	// 使用之前需要先调用EXPECT()，设计模拟场景
	// 注意：设计了几个模拟调用，在使用的是时候就要都用上，而且顺序也要都对上

	userSvc.EXPECT().Signup(gomock.Any(), domain.User{
		Id:    1,
		Email: "12345@qq.com",
	}).Return(errors.New("模拟的错误"))

	// 执行模拟调用
	err := userSvc.Signup(context.Background(), domain.User{
		Id:    1,
		Email: "12345@qq.com",
	})
	t.Log(err)

}
