package web

import (
	"Learn_Go/webook/internal/service"
	"Learn_Go/webook/internal/service/oauth2/wechat"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/lithammer/shortuuid/v4"
	"net/http"
)

type OAuth2WechatHandler struct {
	svc             wechat.Service
	userSvc         service.UserService // 微信登陆也是属于userSvc的服务的
	JWTHandler                          // 这个不用初始化，因为这个结构体，如果是指针的话就需要初始化
	key             []byte
	stateCookieName string
}

func NewOAuth2WechatHandler(svc wechat.Service, userSvc service.UserService) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		svc:             svc,
		userSvc:         userSvc,
		key:             []byte("uF7hZ5sW5fZ7jC1mY1wS9qQ4nQ2gN7lV"),
		stateCookieName: "jwt-state",
		JWTHandler:      newJWTHandler(),
	}
}

func (o *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/oauth2/wechat")
	g.GET("/authurl", o.OAuth2URL)
	g.Any("/callback", o.Callback)

}

func (o *OAuth2WechatHandler) OAuth2URL(ctx *gin.Context) {
	state := uuid.New() // 比较短的uuid，我们在这里就生成好
	val, err := o.svc.AuthURL(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "构造跳转URL失败",
			Code: 5,
		})
		return
	}
	// 在这里设置好 Cookie
	err = o.setStateJWTToken(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "服务器异常",
			Code: 5,
		})
	}
	// 若不返回错误，就拿到构造好的跳转URL，将它传给前端
	ctx.JSON(http.StatusOK, Result{
		Data: val,
	})

}

func (o *OAuth2WechatHandler) Callback(ctx *gin.Context) {
	err := o.verifyState(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "非法请求",
			Code: 4,
		})
		return
	}
	code := ctx.Query("code")
	//state := ctx.Query("state")

	wechatInfo, err := o.svc.VerifyCode(ctx, code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "授权码有误",
			Code: 4,
		})
	}

	// 微信登陆也可能第一次登陆，所以如果是第一次登陆就先注册
	u, err := o.userSvc.FindOrCreateByWechat(ctx, wechatInfo)

	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	// 如果没有返回错误，那么就登陆成功，设置jwtToken
	err = o.setLoginToken(ctx, u.Id)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误！")
		return

	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "登陆成功",
	})
}

func (o *OAuth2WechatHandler) setStateJWTToken(ctx *gin.Context, state string) error {

	sc := StateClaims{ // 初始化 StateClaims
		state: state,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, sc) // 加密 StateClaims
	tokenStr, err := token.SignedString(o.key)             // 这里token是一个结构体，但是传到前端需要一串字符，通过这个方法转换，其中不同的SigningMethod有不同的参数类型

	if err != nil {

		return err
	}
	// 若不返回错误，我们将 JWTToken 存储到 Cookie 中，因为Cookie会自动被带到后端，不需要经过前端，而如果存到Header中，需要前端设置
	ctx.SetCookie(o.stateCookieName, tokenStr, 600, "/oauth2/wechat/callback", "", false, true) // 这里我们将记录了 state 的JWTToken保存在Cookie中，并设置在回调时生成
	return nil
}

func (o *OAuth2WechatHandler) verifyState(ctx *gin.Context) error {
	state := ctx.Query("state") //  从前端获取 state
	ck, err := ctx.Cookie(o.stateCookieName)
	if err != nil {
		return fmt.Errorf("无法获取Cookie %w", err)
	}
	var sc StateClaims
	// 从Cookie中解析出来Token，然后填充 StateClaims
	_, err = jwt.ParseWithClaims(ck, &sc, func(token *jwt.Token) (interface{}, error) {
		return o.key, nil
	})
	if err != nil {
		return fmt.Errorf("解析token失败 %w", err)
	}
	// 判断从 StateCookie 中提取出来的 state 是否和我们设置的一致
	if state != sc.state {
		// state 不匹配，有人搞你
		return fmt.Errorf("state 不匹配")
	}
	return nil

}

// 记录 state 的jwt claims

type StateClaims struct {
	jwt.RegisteredClaims
	state string
}
