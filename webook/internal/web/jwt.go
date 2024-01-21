package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

type JWTHandler struct {
	refreshKey   []byte
	client       redis.Cmdable
	rcExpiration time.Duration
}

func newJWTHandler() JWTHandler {
	return JWTHandler{
		refreshKey:   []byte("uF7hZ5sW5fZ7jC1mY1wS9qQ4nQ2gN7VB"),
		rcExpiration: time.Hour * 24 * 7,
	}
}

func (h *JWTHandler) setLoginToken(ctx *gin.Context, uid int64) error {
	// 生成ssid，这个是长的uuid
	ssid := uuid.New().String()

	// 若没返回错误，则登陆成功，设置JWTToken
	err := h.setRefreshJWTToken(ctx, uid, ssid)
	if err != nil {
		return err

	}
	return h.setJWTToken(ctx, uid, ssid)

}

func (h *JWTHandler) ClearToken(ctx *gin.Context) error {
	// 将前端更新的长短 token 设置为非法值
	ctx.Header("x-jwt-token", "")
	ctx.Header("x-refresh-token", "")
	uc := ctx.MustGet("user").(UserClaims)
	// 这里的过期时间设置为长 token 的过期时间就可以，因为长 token 都过期了，那么检不检测 ssid 都无所谓了
	err := h.client.Set(ctx, fmt.Sprintf("users:ssid:%s", uc.Ssid), "", h.rcExpiration).Err()
	return err
}

// 因为多处需要使用到这个方法，我们把它抽出来，单独放在一个地方，然后在使用到的地方组合它
// 设置短token
func (h *JWTHandler) setJWTToken(ctx *gin.Context, uid int64, ssid string) error {

	uc := UserClaims{ // Claims就表示数据
		Uid:       uid,
		Ssid:      ssid,
		UserAgent: ctx.GetHeader("User-Agent"),
		// 设置过期时间
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30))},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, uc) // SigningMethod的安全性和性能有一些差异，没有要求可随意选
	tokenStr, err := token.SignedString(JWTKey)            // 这里token是一个结构体，但是传到前端需要一串字符，通过这个方法转换，其中不同的SigningMethod有不同的参数类型

	if err != nil {
		return err
	}
	ctx.Header("x-jwt-token", tokenStr) // 希望后端将token放在x-jwt-token里面，前端在请求的Authorization头部带上Bearer token
	return nil
}

// 设置长token
func (h *JWTHandler) setRefreshJWTToken(ctx *gin.Context, uid int64, ssid string) error {
	rc := RefreshClaims{
		Uid:  uid,
		Ssid: ssid,
		// 设置过期时间，长token设置为7天
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.rcExpiration)),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS512, rc)
	refreshTokenStr, err := refreshToken.SignedString(h.refreshKey)
	if err != nil {
		return err
	}
	// 要记得在处理跨域请求那里加上这个header
	ctx.Header("x-refresh-token", refreshTokenStr)
	return nil

}

func ExtractToken(ctx *gin.Context) string {
	// 根据约定，token 在Authorization 头部
	// Bearer XXXXXX
	authCode := ctx.GetHeader("Authorization")
	if authCode == "" {
		return authCode
	}
	segs := strings.Split(authCode, " ") // 按照空格切割字符串，头部分为两部分 Bearer XXXXXX

	if len(segs) != 2 { // 如果不是两部分，则不对
		return ""
	}
	tokenStr := segs[1]
	return tokenStr

}

type RefreshClaims struct {
	jwt.RegisteredClaims
	Ssid string
	Uid  int64
}

type UserClaims struct {
	jwt.RegisteredClaims // 正常就这么使用
	Ssid                 string
	Uid                  int64
	UserAgent            string
}

var JWTKey = []byte("uF7hZ5sW5fZ7jC1mY1wS9qQ4nQ2gN7lJ")
