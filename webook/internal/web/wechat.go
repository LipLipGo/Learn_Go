package web

import (
	"Learn_Go/webook/internal/service"
	"Learn_Go/webook/internal/service/oauth2/wechat"
	"github.com/gin-gonic/gin"
	"net/http"
)

type OAuth2WechatHandler struct {
	svc        wechat.Service
	userSvc    service.UserService // 微信登陆也是属于userSvc的服务的
	JWTHandler                     // 这个不用初始化，因为这个结构体，如果是指针的话就需要初始化
}

func NewOAuth2WechatHandler(svc wechat.Service, userSvc service.UserService) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		svc:     svc,
		userSvc: userSvc,
	}
}

func (o *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/oauth2/wechat")
	g.GET("/authurl", o.OAuth2URL)
	g.Any("/callback", o.Callback)

}

func (o *OAuth2WechatHandler) OAuth2URL(ctx *gin.Context) {
	val, err := o.svc.AuthURL(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "构造跳转URL失败",
			Code: 5,
		})

	}
	// 若不返回错误，就拿到构造好的跳转URL，将它传给前端
	ctx.JSON(http.StatusOK, Result{
		Data: val,
	})

}

func (o *OAuth2WechatHandler) Callback(ctx *gin.Context) {
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
	o.setJWTToken(ctx, u.Id)
	ctx.JSON(http.StatusOK, Result{
		Msg: "登陆成功",
	})

}
