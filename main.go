package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"log"
	_redis "parser/helpers/redis"
	"parser/services"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
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
	vacancies, err := vacancyService.GetVacancies(lastUpdateTime)
	if err != nil {
		panic(err)
	}

	var newVacanciesCount = 0
	for _, vacancy := range vacancies.Vacancies {
		isUnique, err := vacancyService.SaveVacancy(vacancy)
		if err != nil {
			log.Fatal(err)

			continue
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
