package main

import (
	"errors"
	_ "github.com/go-sql-driver/mysql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"log"
	"os"
	_redis "parser/helpers/redis"
	"parser/services"
	"strconv"
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

	lastUpdateTime := redisClient.GetRedisTimeStamp()
	vacancies, err := vacancyService.GetVacancies(lastUpdateTime)
	if err != nil {
		panic(err)
	}

	var newVacanciesCount int = 0
	for _, vacancy := range vacancies.Vacancies {
		isUnique, err := vacancyService.SaveVacancy(vacancy)
		if err != nil {
			log.Fatal(err)
		}

		if isUnique {
			newVacanciesCount++
			// записать в мапу уже готовое сообщение для бота
		}
	}

	if newVacanciesCount != 0 {
		redisClient.SetRedisTimeStamp()

		// кидаем уведомление боту
		telegramToken, isExists := os.LookupEnv("TELEGRAM_BOT_TOKEN")
		if !isExists {
			panic(errors.New("telegram token not found"))
		}

		telegramChatID, isExists := os.LookupEnv("TELEGRAM_CHAT_ID")
		if !isExists {
			panic(errors.New("telegram chat id not found"))
		}
		bot, err := tgbotapi.NewBotAPI(telegramToken)
		if err != nil {
			panic(err)
		}

		chatId, _ := strconv.Atoi(telegramChatID)
		msg := tgbotapi.NewMessage(int64(chatId), "aaaa")
		bot.Send(msg)
	}
}
