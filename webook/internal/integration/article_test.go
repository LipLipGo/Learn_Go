package integration

import (
	"Learn_Go/webook/internal/integration/startup"
	"Learn_Go/webook/internal/repository/dao"
	ijwt "Learn_Go/webook/internal/web/jwt"
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

// 可以使用套件改写这种代码组织方式

func TestArticleHandler_Edit(t *testing.T) {
	server := gin.Default()
	artHdl := startup.InitArticleHandler()
	server.Use(func(ctx *gin.Context) {
		ctx.Set("user", ijwt.UserClaims{Uid: 123})
	})
	artHdl.RegisterRouters(server)
	db := startup.InitDB()
	testCases := []struct {
		name string
		// 提前准备数据
		before func(t *testing.T)
		// 测试后删除数据
		after func(t *testing.T)

		// 前端传过来数据，JSON
		art Article

		wantCode int
		wantRes  Result[int64]
	}{
		{ // 这里有一个问题，要编辑帖子，需要登录态，那么就需要模拟一个登录态（将 JWT token 放到 Header 中），或者在 wire 初始化中将登录功能去掉，模拟一个 UserId
			name: "新建帖子",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
				// 验证，数据保存到数据库中，需要先定义表结构
				var art dao.Article
				err := db.Where("author_id=?", 123).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				art.Utime = 0
				art.Ctime = 0
				assert.Equal(t, dao.Article{
					Id:       1,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
				}, art)
				db.Exec("truncate table `articles`")

			},
			art: Article{
				Title:    "我的标题",
				Content:  "我的内容",
				AuthorId: 123,
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				// 希望拿到的文章 ID 是 1
				Msg:  "保存成功",
				Data: 1,
			},
		},
		{
			name: "修改帖子",
			before: func(t *testing.T) {
				err := db.Create(dao.Article{
					Id:       2,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
					Ctime:    456,
					Utime:    789,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 验证，数据保存到数据库中，需要先定义表结构
				var art dao.Article
				err := db.Where("id=?", 2).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Utime > 789)
				art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       2,
					Title:    "新的标题",
					Content:  "新的内容",
					AuthorId: 123,
					Ctime:    456,
				}, art)
				db.Exec("truncate table `articles`")

			},
			art: Article{
				Id:      2,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Msg:  "保存成功",
				Data: 2,
			},
		},
		{
			// 模拟攻击者修改前端传下来的帖子 Id
			name: "修改帖子 -- 修改别人的帖子",
			before: func(t *testing.T) {
				err := db.Create(dao.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
					Ctime:    456,
					Utime:    789,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 验证，数据没有变
				var art dao.Article
				err := db.Where("id=?", 3).First(&art).Error
				assert.NoError(t, err)
				assert.Equal(t, dao.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
					Ctime:    456,
					Utime:    789,
				}, art)
				db.Exec("truncate table `articles`")

			},
			art: Article{
				Id:      2,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)

			art, err := json.Marshal(tc.art)
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost,
				"/articles/edit",
				bytes.NewReader(art))
			req.Header.Set("Content-Type", "application/json")
			assert.NoError(t, err)

			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, req)

			assert.Equal(t, tc.wantCode, recorder.Code)
			if tc.wantCode != http.StatusOK {
				return
			}
			// 反序列化为结构体
			var res Result[int64]
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantRes, res)

		})
	}
}

type Article struct {
	Id       int64  `json:"id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	AuthorId int64  `json:"authorId"`
}

type Result[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"` // 文章 ID
}
