package redis

import (
	"github.com/go-redis/redis"
	"log"
)

const prefix = "goffer:cache:"

var client *redis.Client

func init() {
	client = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := client.Ping().Result()
	log.Println(pong, err)
}

func SetCache(key string, value interface{}) error {
	return client.Set(prefix+key, value, 0).Err()
}

func GetCache(key string) (string, error) {
	val, err := client.Get(prefix + key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}
