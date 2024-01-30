package web

import (
	"Learn_Go/webook/internal/domain"
	"Learn_Go/webook/internal/service"
	svcmocks "Learn_Go/webook/internal/service/mocks"
	ijwt "Learn_Go/webook/internal/web/jwt"
	"Learn_Go/webook/pkg/logger"
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestArticleHandler_Publish(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) service.ArticleService
		reqBody  string
		wantCode int
		wantRes  Result
	}{
		{
			name: "新建并发表成功",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				return svc
			},
			reqBody: `
{
	"title" :"我的标题",
	"content" : "我的内容"
}
`,
			wantCode: http.StatusOK,
			wantRes: Result{
				Msg:  "发表成功",
				Data: float64(1),
			},
		},
		{
			name: "更新并发表成功",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				return svc
			},
			reqBody: `
{
	"id": 1,
	"title" :"我的标题",
	"content" : "我的内容"
}
`,
			wantCode: http.StatusOK,
			wantRes: Result{
				Msg:  "发表成功",
				Data: float64(1),
			},
		},
		{
			name: "更新但发表失败",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), errors.New("发表失败"))
				return svc
			},
			reqBody: `
{
	"id": 1,
	"title" :"我的标题",
	"content" : "我的内容"
}
`,
			wantCode: http.StatusOK,
			wantRes: Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},
		{
			name: "Bind出错",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				return svc
			},
			reqBody: `
{
	"title" :"我的标题",hkjhk
	"content" : "我的内容"
}
`,
			wantCode: http.StatusOK,
			wantRes: Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			artSvc := tc.mock(ctrl)
			artHdl := NewArticleHandler(artSvc, logger.NewNopLogger())

			server := gin.Default()
			// 这里需要一个登录态
			server.Use(func(ctx *gin.Context) {
				ctx.Set("user", ijwt.UserClaims{
					Uid: 123,
				})
			})
			artHdl.RegisterRouters(server)

			req, err := http.NewRequest(http.MethodPost,
				"/articles/publish",
				bytes.NewReader([]byte(tc.reqBody)))
			req.Header.Set("Content-Type", "application/json")
			assert.NoError(t, err)

			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, req)

			if recorder.Code == 400 {
				return
			}
			var res Result
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}
