package main

import (
	"errors"
	_ "github.com/go-sql-driver/mysql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"log"
	"os"
	"parser/db/models"
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
	//redisClient, err := _redis.NewRedisClient()
	//if err != nil {
	//	panic(err)
	//}
	//
	vacancyService, err := services.CreateVacancyService()
	if err != nil {
		panic(err)
	}
	//
	//lastUpdateTime := redisClient.GetRedisTimeStamp()
	//vacancies, err := vacancyService.GetVacancies(lastUpdateTime)
	//if err != nil {
	//	panic(err)
	//}
	//
	//var newVacanciesCount int = 0
	//for _, vacancy := range vacancies.Vacancies {
	//	isUnique, err := vacancyService.SaveVacancy(vacancy)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//
	//	if isUnique {
	//		newVacanciesCount++
	//		//err := sendToChannel(prepareMessage(&vacancy))
	//		//if err != nil {
	//		//	panic(err)
	//		//}
	//	}
	//}
	//
	//if newVacanciesCount != 0 {
	//	redisClient.SetRedisTimeStamp()
	//}

	vacancy, _ := vacancyService.GetVacancyById(26673199)
	msg := prepareMessage(vacancy)
	err = sendToChannel(msg)
	if err != nil {
		panic(err)
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

	message.ParseMode = tgbotapi.ModeHTML
	_, err = bot.Send(message)
	if err != nil {
		log.Fatal(err)

		return err
	}

	return nil
}

func prepareMessage(vacancy *models.Vacancy) string {
	var str strings.Builder
	str.WriteString("<b>" + vacancy.Name)
	str.WriteString(" в " + vacancy.Employer.Name + "</b>\n")

	if vacancy.Salary.From != 0 && vacancy.Salary.To == 0 {
		str.WriteString("Зарплата: от " + strconv.Itoa(vacancy.Salary.From) + "\n")
	} else if vacancy.Salary.To != 0 && vacancy.Salary.From == 0 {
		str.WriteString("Зарплата: до " + strconv.Itoa(vacancy.Salary.To) + "\n")
	} else if vacancy.Salary.From != 0 && vacancy.Salary.To != 0 {
		str.WriteString("Зарплата: " +
			"от " + strconv.Itoa(vacancy.Salary.From) +
			" до " + strconv.Itoa(vacancy.Salary.To) +
			"\n",
		)
	}

	if vacancy.Snippet.Description != "" {
		str.WriteString("\n" + vacancy.Snippet.Description + "\n")
	}

	if vacancy.Snippet.Requirements != "" {
		str.WriteString("Требования: " + vacancy.Snippet.Requirements + "\n")
	}

	str.WriteString("<b><a href='" +
		vacancy.URL +
		"'>Откликнуться</a></b>")

	return str.String()
}
