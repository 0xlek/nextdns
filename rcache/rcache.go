package rcache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/nextdns/nextdns/resolver"
	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
	maxAge time.Duration
}

func NewCache(addr string, maxAge time.Duration) *Cache {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	pong, err := client.Ping(context.Background()).Result()

	fmt.Println(pong, err)
	if err != nil {
		panic(errors.New("redis failed to connect:"))
	}

	cache := &Cache{
		client: client,
		maxAge: maxAge,
	}

	return cache
}

func (c *Cache) Ping(ctx context.Context) error {
	pong, err := c.client.Ping(ctx).Result()
	if err != nil {
		return err
	}
	fmt.Println(pong)
	return nil
}

func (c *Cache) Add(key resolver.CacheKey, value *resolver.CacheValue) {
	redisKey := key.String()

    record, _ := c.client.Get(context.Background(), redisKey).Result()

    if record != "" {
        return
    }

	b, err := json.Marshal(value)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(string(b))
	err = c.client.Set(context.Background(), redisKey, string(b), c.maxAge).Err()

	if err != nil {
		fmt.Println(err)
	}
}

func (c *Cache) Get(key resolver.CacheKey) (resolver.CacheValue, bool) {
	redisKey := key.String()
	result, err := c.client.Get(context.Background(), redisKey).Result()

	if err != nil {
		return resolver.CacheValue{}, false
	}

	var response resolver.CacheValue

	err = json.Unmarshal([]byte(result), &response)

	if err != nil {
		fmt.Println(err, "UNMARSHALL")
	}

	return response, true
}

func (c *Cache) Delete(key resolver.CacheKey) bool {
	redisKey := key.String()
	err := c.client.Del(context.Background(), redisKey).Err()

	return err != nil
}

func (c *Cache) Keys() []string {
	keys, err := c.client.Keys(context.Background(), "*").Result()
	if err != nil {
		return keys
	}
	return keys
}
