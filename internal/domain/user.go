package domain

import "time"

// User 领域对象，是 DDD 中的 entity
// BO(business object)
type User struct {
	Id         int64
	Email      string
	Password   string
	Phone      string
	Nickname   string
	WechatInfo WechatInfo // 没有组合，因为其他oauth2第三方，例如钉钉可能也有OpenID，会导致重名
	Ctime      time.Time
}
