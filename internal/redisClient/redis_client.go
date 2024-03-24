package redisClient

import "github.com/redis/go-redis/v9"

var Rdb *redis.Client

func InitRedis() *redis.Client {
	Rdb = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})
	return Rdb
}
