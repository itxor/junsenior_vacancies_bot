package telegram

import (
	"errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"os"
)

type TelegramBot struct {
	client *tgbotapi.BotAPI
}

func InitTelegramBot() (*TelegramBot, error) {
	telegramToken, isExists := os.LookupEnv("TELEGRAM_BOT_TOKEN")
	if !isExists {
		log.Fatalln("telegram token not found")

		return nil, errors.New("telegram token not found")
	}

	bot, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		log.Fatal(err)

		return nil, err
	}

	telegramBot := &TelegramBot{
		client: bot,
	}

	return telegramBot, nil
}

func (tg *TelegramBot) SendHTMLMessage(msg string, channelName string) error {
	message := tgbotapi.NewMessageToChannel(
		channelName,
		msg,
	)

	message.ParseMode = tgbotapi.ModeHTML
	_, err := tg.client.Send(message)
	if err != nil {
		log.Fatal(err)

		return err
	}

	return nil
}
