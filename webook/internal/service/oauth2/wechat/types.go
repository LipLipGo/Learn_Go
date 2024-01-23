package wechat

import (
	"Learn_Go/webook/internal/domain"
	"Learn_Go/webook/pkg/logger"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Service interface {
	AuthURL(ctx context.Context, state string) (string, error)
	VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error)
}

var redirectURL = url.PathEscape("https://meoying.com/oauth2/wechat/callback") // 这里需要对这个URL进行转义

type service struct {
	appId     string // 一般来说是固定的
	appSecret string
	client    *http.Client
	l         logger.LoggerV1
}

func NewWechatService(appId string, appSecret string, l logger.LoggerV1) Service {
	return &service{
		appId:     appId,
		client:    http.DefaultClient, // 这里使用这个，是因为很少会去改这个，如果后面有需求，可以做成依赖注入的形式
		appSecret: appSecret,
		l:         l,
	}
}

func (s *service) AuthURL(ctx context.Context, state string) (string, error) {
	const AuthURLPattern = `https://open.weixin.qq.com/connect/qrconnect?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_login&state=%s#wechat_redirect`
	return fmt.Sprintf(AuthURLPattern, s.appId, redirectURL, state), nil
}

func (s *service) VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error) {
	accessTokenUrl := fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code", s.appId, s.appSecret, code)
	// 构造一个请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, accessTokenUrl, nil)
	if err != nil {
		return domain.WechatInfo{}, err
	}
	httpRes, err := s.client.Do(req)
	if err != nil {
		return domain.WechatInfo{}, err
	}
	// 因为这里返回的是一个json的格式，而我们是定义了一个结构体来存放返回的信息，所以在这里反序列化它
	var res Result
	err = json.NewDecoder(httpRes.Body).Decode(&res) // 这里反序列化使用NewDecoder，是因为响应的body是readCloser,而不是字符串，无法使用Unmarshal
	if err != nil {
		// 反序列化出错
		return domain.WechatInfo{}, err
	}
	// 返回的响应错误码，如果不为0，就是调用失败，具体错误码信息可以查看开发文档
	if res.ErrCode != 0 {
		return domain.WechatInfo{}, fmt.Errorf("调用微信接口失败 errcode %d, errmsg %s", res.ErrCode, res.ErrMsg)
	}

	return domain.WechatInfo{
		UnionId: res.UnionId,
		OpenId:  res.OpenId,
	}, nil

}

type Result struct {
	// 接口调用凭证
	AccessToken string `json:"access_token"`
	// access_token接口调用凭证超时时间，单位（秒）
	ExpiresIn int64 `json:"expires_in"`
	// 用户刷新access_token
	RefreshToken string `json:"refresh_token"`
	// 授权用户唯一标识
	OpenId string `json:"openid"`
	// 用户授权的作用域，使用逗号（,）分隔
	Scope string `json:"scope"`
	// 当且仅当该网站应用已获得该用户的userinfo授权时，才会出现该字段。
	UnionId string `json:"unionid"`

	// 错误返回
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}
