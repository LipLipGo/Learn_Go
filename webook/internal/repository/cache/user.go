package cache

import (
	"Learn_Go/webook/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

var ErrKeyNotExist = redis.Nil

type UserCache interface {
	Get(ctx context.Context, uid int64) (domain.User, error)
	Set(ctx context.Context, du domain.User) error
}

type RedisUserCache struct {
	cmd        redis.Cmdable // 操作Redis的应用，为什么不适用client，因为client是具体的实现，而Cmdable是面向接口编程，扩展性更好
	expiration time.Duration // 过期时间
}

func (c *RedisUserCache) Get(ctx context.Context, uid int64) (domain.User, error) {
	key := c.Key(uid)
	// 假定这个地方使用 JSON 序列化
	data, err := c.cmd.Get(ctx, key).Result()
	if err != nil {
		return domain.User{}, err
	}
	var u domain.User
	// 反序列化
	err = json.Unmarshal([]byte(data), &u)
	return u, err
}

func (c *RedisUserCache) Key(uid int64) string {
	return fmt.Sprintf("user:info:%d", uid)
}

func (c *RedisUserCache) Set(ctx context.Context, du domain.User) error {
	key := c.Key(du.Id)
	// 序列化
	data, err := json.Marshal(du)
	if err != nil {
		return err
	}

	err = c.cmd.Set(ctx, key, data, c.expiration).Err()
	return err

}

func NewRedisUserCache(cmd redis.Cmdable) UserCache {
	return &RedisUserCache{
		cmd:        cmd,              // 从外面传，不要自己去初始化需要的东西
		expiration: time.Minute * 15, // 过期时间可以直接写死
	}
}
