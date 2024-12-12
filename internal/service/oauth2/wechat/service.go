package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"red-feed/internal/domain"
	"red-feed/pkg/logger"
)

const (
	redirectURI = "http://localhost:8080/oauth2/wechat/callback" // 配置为实际在微信开放平台中配置的域名
)

type Service interface {
	AuthURL(ctx context.Context, state string) (string, error)
	VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error)
}

type service struct {
	appId     string
	appSecret string
	client    *http.Client
	l         logger.Logger
}

func NewService(appId string, appSecret string, l logger.Logger) Service {
	return &service{
		appId:     appId,
		appSecret: appSecret,
		l:         l,
	}
}

func (s *service) AuthURL(ctx context.Context, state string) (string, error) {
	// 构造url模板
	const urlPattern = "https://open.weixin.qq.com/connect/qrconnect?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_login&state=%s#wechat_redirect"
	return fmt.Sprintf(urlPattern, s.appId, redirectURI, state), nil
}

func (s *service) VerifyState(ctx context.Context, state string) error {
	return nil
}

func (s *service) VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error) {
	// 根据code和state尝试去请求微信换取access_token等信息
	const targetPattern = "https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code"
	target := fmt.Sprintf(targetPattern, s.appId, s.appSecret, code)
	// 构造请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return domain.WechatInfo{}, fmt.Errorf("构造请求失败，%w", err)
	}
	// 发送请求
	resp, err := s.client.Do(req)
	if err != nil {
		return domain.WechatInfo{}, fmt.Errorf("发送请求失败%w", err)
	}
	// 获得响应
	decoder := json.NewDecoder(resp.Body)
	var res Result
	err = decoder.Decode(&res)
	if err != nil {
		return domain.WechatInfo{}, err
	}
	if res.ErrCode != 0 {
		return domain.WechatInfo{},
			fmt.Errorf("微信返回错误响应，错误码：%d，错误信息：%s", res.ErrCode, res.ErrMsg)
	}

	return domain.WechatInfo{
		OpenID:  res.OpenID,
		UnionID: res.UnionID,
	}, nil
}

// Result 通过code和state请求微信返回的数据
type Result struct {
	ErrCode int64  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`

	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`

	OpenID  string `json:"openid"`
	Scope   string `json:"scope"`
	UnionID string `json:"unionid"`
}
