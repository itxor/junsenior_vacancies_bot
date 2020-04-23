package redis_helper

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"log"
	"os"
	"strconv"
	"time"
)

type Redis struct {
	client *redis.Client
}

func NewRedisClient() (*Redis, error) {
	redisUrl, exists := os.LookupEnv("REDIS_URL")
	if !exists {
		log.Fatalln("redis settings not found")

		return nil, errors.New("redis setting not found")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     redisUrl,
		Password: "",
		DB:       0,
	})
	Redis := &Redis{client: client}

	_, err := Redis.client.Ping().Result()
	if err != nil {
		log.Fatal(err)

		return nil, err
	}

	return Redis, nil
}

func (redis *Redis) GetRedisTimeStamp() string {
	lastUpdateTime, err := redis.client.Get("last-vacancies-update").Result()
	if err != nil && err.Error() != "redis: nil" {
		panic(err)
	}

	return lastUpdateTime
}

func (redis *Redis) SetRedisTimeStamp() {
	err := redis.client.Set("last-vacancies-update", getCurrentDateStr(), 0).Err()
	if err != nil {
		panic(err)
	}
}

func getCurrentDateStr() string {
	year, month, day := time.Now().Date()
	var intMonth = int(month)

	return strconv.Itoa(year) + "-" + fmt.Sprintf("%02d", intMonth) + "-" + strconv.Itoa(day)
}
