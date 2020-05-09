package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"log"
	"os"
	_redis "parser/helpers/redis"
	"parser/services"
	"time"
)

func init() {
	t := time.Now()
	formatted := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
	fmt.Println("Start time: " + formatted)

	path := "/root/go/src/github.com/itxor/junsenior_vacancies_bot/.env"
	if _, err := os.Stat(path); err != nil {
		log.Println("No .env file found")
		panic(err)
	}

	if err := godotenv.Load(
		os.ExpandEnv(path),
	); err != nil {
		log.Println("No .env file load")
		panic(err)
	}
}

func main() {
	redisClient, err := _redis.NewRedisClient()
	if err != nil {
		panic(err)
	}

	vacancyService, err := services.CreateVacancyService()
	if err != nil {
		panic(err)
	}

	telegramService, err := services.CreateTelegramService()
	if err != nil {
		panic(err)
	}

	lastUpdateTime := redisClient.GetRedisTimeStamp()
	if lastUpdateTime == "" {
		lastUpdateTime = time.Now().Format("YYYY-MM-DD")
	}
	fmt.Printf("Redis timestamp: %s\n", lastUpdateTime)

	vacancies, err := vacancyService.GetVacancies(lastUpdateTime)
	if err != nil {
		panic(err)
	}
	fmt.Print(len(vacancies.Vacancies))

	var newVacanciesCount = 0
	for _, vacancy := range vacancies.Vacancies {
		isUnique, err := vacancyService.SaveVacancy(vacancy)
		if err != nil {
			log.Fatal(err)
		}

		if isUnique {
			newVacanciesCount++

			err := telegramService.SendMessageByVacancy(&vacancy)
			if err != nil {
				panic(err)
			}
		}
	}

	if newVacanciesCount != 0 {
		redisClient.SetRedisTimeStamp()
	}
}
