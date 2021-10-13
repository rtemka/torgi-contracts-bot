package bot

import (
	"fmt"
	"log"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	TOKEN string = "2003091653:AAHHuYqtRHcF2HZoHm3wbRUpaMlu2qEnws8"
)

type Config struct {
	Port   string
	AppURL string
}

func Start(c *Config) error {
	bot, err := initBot()
	if err != nil {
		return err
	}

	log.Printf("Telegram	->	Authorized on account	[%s]", bot.Self.UserName)

	msg, err := checkWebhook(bot, c)
	if err != nil {
		return err
	}

	log.Printf("Telegram	->	Webhook cheсk:	[%s]", msg)

	updates := bot.ListenForWebhook("/" + bot.Token)

	go http.ListenAndServe(":"+c.Port, nil)

	for update := range updates {
		// log.Printf("%+v\n", update)
		go handleUpdate(bot, &update)
	}

	return nil
}

func initBot() (*tgbotapi.BotAPI, error) {

	bot, err := tgbotapi.NewBotAPI(TOKEN)
	if err != nil {
		return nil, err
	}

	// bot.Debug = true

	return bot, nil
}

func checkWebhook(bot *tgbotapi.BotAPI, c *Config) (string, error) {

	info, err := bot.GetWebhookInfo()

	if err != nil {
		return "", err
	}

	if info.LastErrorDate != 0 {
		return "", fmt.Errorf("telegram callback failed: %s", info.LastErrorMessage)
	}

	if info.IsSet() {
		return "webhook already set and available", err
	}

	_, err = bot.SetWebhook(tgbotapi.NewWebhook(c.AppURL + "/" + bot.Token))
	if err != nil {
		return "", err
	}

	return "new webhook installed", nil
}

func handleUpdate(bot *tgbotapi.BotAPI, u *tgbotapi.Update) {
	if u.Message.IsCommand() {
		msg := tgbotapi.NewMessage(u.Message.Chat.ID, "")
		switch u.Message.Command() {
		case "help":
			msg.Text = "Напиши /hi или /status."
		case "hi":
			msg.Text = "Привет :)"
		case "status":
			msg.Text = "Все ок!"
		case "withArgument":
			msg.Text = "Ты добавил к команде аргумент: " + u.Message.CommandArguments()
		default:
			msg.Text = "Не знаю такой команды"
		}
		bot.Send(msg)
	}
}
