package svc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStore 实现了 base64Captcha.Store 接口
// 使用 Redis 作为验证码的存储后端
// https://github.com/mojocn/base64Captcha/blob/master/README.md
type RedisStore interface {
	// Set sets the digits for the captcha id.
	Set(id string, value string) error

	// Get returns stored digits for the captcha id. Clear indicates
	// whether the captcha must be deleted from the store.
	Get(id string, clear bool) string

	//Verify captcha's answer directly
	Verify(id, answer string, clear bool) bool
}

// 创建 RedisStore 实例
func NewRedisStore(
	ctx context.Context,
	client redis.UniversalClient,
	prefix string,
	ttl time.Duration,
) RedisStore {
	return &redisStore{
		Ctx:    ctx,
		Client: client,
		Prefix: prefix,
		TTL:    ttl,
	}
}

type redisStore struct {
	Ctx    context.Context
	Client redis.UniversalClient
	Prefix string
	TTL    time.Duration
}

func (rs *redisStore) Set(id string, value string) error {
	key := rs.Prefix + id
	err := rs.Client.Set(rs.Ctx, key, strings.ToLower(value), rs.TTL).Err()
	if err != nil {
		return fmt.Errorf("failed to set captcha in redis: %w", err)
	}

	return nil
}

func (rs *redisStore) Get(id string, clear bool) string {
	key := rs.Prefix + id
	val, err := rs.Client.Get(rs.Ctx, key).Result()
	if err != nil {
		return ""
	}

	if clear {
		err := rs.Client.Del(rs.Ctx, key).Err()
		if err != nil {
			_ = err
		}
	}

	return strings.ToLower(val)
}

func (rs *redisStore) Verify(id, answer string, clear bool) bool {
	stored := rs.Get(id, clear)
	return stored == strings.ToLower(answer)
}
