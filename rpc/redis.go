package rpc

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"php-thrift-go-server/client"
	"php-thrift-go-server/util"
)

//能否ping通Redis
func Ping()  {
	pong, err := client.RedisClient.Ping().Result()
	fmt.Println(pong,err)
}

//key:value
func RedisSet(key string, value interface{}) error {
	str := util.JsonString(value)
	err := client.RedisClient.Set(key, str, 0).Err()
	return err
}

func RedisGet(key string) (val string, err error) {
	val, err = client.RedisClient.Get(key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key: %s not exist", key)
	} else if err != nil {
		return "", errors.New("redis internal error")
	} else {
		return
	}

}