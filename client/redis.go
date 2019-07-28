package client

import (
	"fmt"
	"github.com/go-redis/redis"
	"php-thrift-go-server/conf"
)

var RedisClient redis.Client

func InitRedis(config conf.RedisConf)  {
	RedisClient = *redis.NewClient(&redis.Options{
		Addr:config.Addr,
		Password:"",
		DB:0,
	})
	pong, err := RedisClient.Ping().Result()
	fmt.Println(pong, "========", err)
}
