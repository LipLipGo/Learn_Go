package jwt

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"net/http"
	"strings"
	"time"
)

type RedisJWTHandler struct {
	client       redis.Cmdable
	rcExpiration time.Duration
}

func NewRedisJWTHandler(client redis.Cmdable) Handler {
	return &RedisJWTHandler{
		client:       client,
		rcExpiration: time.Hour * 24 * 7,
	}
}

func (h *RedisJWTHandler) CheckSession(ctx *gin.Context, ssid string) error {
	// 在这里去校验 ssid 是否失效，因为我们可以在前面先校验一下 token ，避免无效的查询 Redis

	cnt, err := h.client.Exists(ctx, fmt.Sprintf("users:ssid:%s", ssid)).Result()

	if err != nil {
		return err
	}

	// 这种判定方式过于严格，因为一旦 redis 崩溃了，就无法继访问服务
	if cnt > 0 {
		// ssid 无效或者 redis 有问题
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return errors.New("token 无效")
	}
	return nil
}

func (h *RedisJWTHandler) ExtractToken(ctx *gin.Context) string {
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

func (h *RedisJWTHandler) SetLoginToken(ctx *gin.Context, uid int64) error {
	// 生成ssid，这个是长的uuid
	ssid := uuid.New().String()

	// 若没返回错误，则登陆成功，设置JWTToken
	err := h.setRefreshJWTToken(ctx, uid, ssid)
	if err != nil {
		return err

	}
	return h.SetJWTToken(ctx, uid, ssid)

}

func (h *RedisJWTHandler) ClearToken(ctx *gin.Context) error {
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
func (h *RedisJWTHandler) SetJWTToken(ctx *gin.Context, uid int64, ssid string) error {

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
func (h *RedisJWTHandler) setRefreshJWTToken(ctx *gin.Context, uid int64, ssid string) error {
	rc := RefreshClaims{
		Uid:  uid,
		Ssid: ssid,
		// 设置过期时间，长token设置为7天
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.rcExpiration)),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS512, rc)
	refreshTokenStr, err := refreshToken.SignedString(RefreshJWTKey)
	if err != nil {
		return err
	}
	// 要记得在处理跨域请求那里加上这个header
	ctx.Header("x-refresh-token", refreshTokenStr)
	return nil

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
var RefreshJWTKey = []byte("uF7hZ5sW5fZ7jC1mY1wS9qQ4nQ2gN7VB")
