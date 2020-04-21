package redis_helper

import (
	"fmt"
	"github.com/go-redis/redis"
	"log"
	"strconv"
	"time"
)

var redisClient *redis.Client

func init() {
	redisClient, err := NewRedisClient()
	if err != nil {
		panic(err)
	}
}

func NewRedisClient() (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	_, err := client.Ping().Result()
	if err != nil {
		log.Fatal(err)

		return nil, err
	}

	return client, nil
}

func GetRedisTimeStamp() string {
	lastUpdateTime, err := redisClient.Get("last-vacancies-update").Result()
	if err != nil && err.Error() != "redis: nil" {
		panic(err)
	}

	return lastUpdateTime
}

func SetRedisTimeStamp() {
	err := redisClient.Set("last-vacancies-update", getCurrentDateStr(), 0).Err()
	if err != nil {
		panic(err)
	}
}

func getCurrentDateStr() string {
	year, month, day := time.Now().Date()
	var intMonth int = int(month)

	return strconv.Itoa(year) + "-" + fmt.Sprintf("%02d", intMonth) + "-" + strconv.Itoa(day)
}

