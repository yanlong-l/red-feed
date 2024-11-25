package auth

import (
	"context"
	"red-feed/internal/service/sms"

	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
	svc sms.Service
	key string
}

func (s *Service) GenSmsToken(ctx context.Context, tpl string) string {
	c := Claims{
		Tpl: tpl,
	}
	jwt.NewWithClaims(jwt.SigningMethodES256, &c)
	token, err := jwt.NewWithClaims(jwt.SigningMethodES256, &c).SignedString([]byte(s.key))
	if err != nil {
		return ""
	}
	return token
}

// 其中tpl是线下业务方申请的一个token，token中预期可以解析出对应的模板ID
func (s *Service) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	var tc Claims // token claims
	jwt.ParseWithClaims(tpl, &tc, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.key), nil
	})
	return s.svc.Send(ctx, tpl, args, numbers...)
}

type Claims struct {
	jwt.RegisteredClaims
	Tpl string
}
