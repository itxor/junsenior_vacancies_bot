package main

import (
	"errors"
	_ "github.com/go-sql-driver/mysql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"log"
	"os"
	"parser/db/models"
	_redis "parser/helpers/redis"
	"parser/services"
	"strconv"
	"strings"
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
			err := sendToChannel(prepareMessage(&vacancy))
			if err != nil {
				panic(err)
			}
		}
	}

	if newVacanciesCount != 0 {
		redisClient.SetRedisTimeStamp()
	}
}

func sendToChannel(msg string) error {
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
	message := tgbotapi.NewMessage(
		int64(chatId),
		msg,
	)
	_, err = bot.Send(message)
	if err != nil {
		log.Fatal(err)

		return err
	}

	return nil
}

func prepareMessage(vacansy *models.Vacancy) string {
	var str strings.Builder
	str.WriteString(vacansy.Name + "\n")

	if vacansy.Salary.From == 0 && vacansy.Salary.To == 0 {
		str.WriteString("Зарплата: не указана\n")
	} else if vacansy.Salary.From != 0 && vacansy.Salary.To == 0 {
		str.WriteString("Зарплата: от " + strconv.Itoa(vacansy.Salary.From) + "\n")
	} else if vacansy.Salary.To != 0 && vacansy.Salary.From == 0 {
		str.WriteString("Зарплата: до " + strconv.Itoa(vacansy.Salary.To) + "\n")
	} else {
		str.WriteString("Зарплата: " +
			"от " + strconv.Itoa(vacansy.Salary.From) +
			" до " + strconv.Itoa(vacansy.Salary.To) +
			"\n",
		)
	}

	str.WriteString(vacansy.Snippet.Description + "\n")
	str.WriteString("Требования: " + vacansy.Snippet.Requirements + "\n")
	str.WriteString(vacansy.Employer.Name + "\n")

	return str.String()
}
