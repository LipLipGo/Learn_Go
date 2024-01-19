package web

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
)

type JWTHandler struct {
}

// 因为多处需要使用到这个方法，我们把它抽出来，单独放在一个地方，然后在使用到的地方组合它
func (h *JWTHandler) setJWTToken(ctx *gin.Context, uid int64) {

	uc := UserClaims{ // Claims就表示数据
		Uid:       uid,
		UserAgent: ctx.GetHeader("User-Agent"),
		// 设置过期时间
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30))},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, uc) // SigningMethod的安全性和性能有一些差异，没有要求可随意选
	tokenStr, err := token.SignedString(JWTKey)            // 这里token是一个结构体，但是传到前端需要一串字符，通过这个方法转换，其中不同的SigningMethod有不同的参数类型

	if err != nil {
		ctx.String(http.StatusOK, "系统错误！")
	}
	ctx.Header("x-jwt-token", tokenStr) // 希望后端将token放在x-jwt-token里面，前端在请求的Authorization头部带上Bearer token

}

type UserClaims struct {
	jwt.RegisteredClaims // 正常就这么使用
	Uid                  int64
	UserAgent            string
}

var JWTKey = []byte("uF7hZ5sW5fZ7jC1mY1wS9qQ4nQ2gN7lJ")
