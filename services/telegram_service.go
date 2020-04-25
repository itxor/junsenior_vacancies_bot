package services

import (
	"errors"
	"log"
	"os"
	"parser/db/models"
	"parser/helpers/telegram"
	"strconv"
	"strings"
)

type TelegramService struct {
	bot *telegram.TelegramBot
	chatId int64
}

func CreateTelegramService() (*TelegramService, error) {
	bot, err := telegram.InitTelegramBot()
	if err != nil {
		log.Fatal(err)

		return nil, err
	}

	telegramChatID, isExists := os.LookupEnv("TELEGRAM_CHAT_ID")
	if !isExists {
		log.Fatal(errors.New("telegram chat id not found"))

		return nil, err
	}

	tempId, _ := strconv.Atoi(telegramChatID)
	id := int64(tempId)

	service := &TelegramService{
		bot: bot,
		chatId: id,
	}

	return service, nil
}

func (ts *TelegramService) SendMessageByVacancy(vacancy *models.Vacancy) error {
	msg := prepareMessage(vacancy)
	err := ts.bot.SendHTMLMessage(msg, ts.chatId)
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
		str.WriteString("\nТребования: " + vacancy.Snippet.Requirements + "\n")
	}

	str.WriteString("<b><a href='" +
		vacancy.URL +
		"'>Откликнуться</a></b>")

	return str.String()
}
