package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"red-feed/internal/domain"
	"time"
)

type UserCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

func NewUserCache(client redis.Cmdable) *UserCache {
	return &UserCache{
		client:     client,
		expiration: time.Minute * 60,
	}
}

func (uc *UserCache) Set(ctx context.Context, u domain.User) error {
	val, err := json.Marshal(u)
	if err != nil {
		return err
	}
	return uc.client.Set(ctx, uc.Key(u.Id), val, uc.expiration).Err()
}

func (uc *UserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	var user domain.User
	res := uc.client.Get(ctx, uc.Key(id))
	if res.Err() != nil {
		return user, res.Err()
	}
	userBytes, err := res.Bytes()
	if err != nil {
		return user, err
	}
	err = json.Unmarshal(userBytes, &user)
	if err != nil {
		return user, err
	}
	return user, nil
}

func (uc *UserCache) Key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}
