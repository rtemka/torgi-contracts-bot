package bot

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	botDB "trbot/src/botDB"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Config is the settigns
// for the bot instance
type Config struct {
	BotName       string
	Port          string
	AppURL        string
	BotToken      string
	DbUpdateToken string
	DB            *sql.DB
}

// bot is the main controller
// over the application logic
type bot struct {
	name string
	api  *tgbotapi.BotAPI
	dbh  dbUpdateHandler
	tgh  tgUpdateHandler
}

// databaseHandler is the logic responsible
// for processing incoming updates for the database
type dbUpdateHandler interface {
	http.Handler
}

// tgUpdateHandler is the logic responsible for
// processing incoming telegram updates i.e.
// responding to the users messages
type tgUpdateHandler interface {
	handleUpdate(api *tgbotapi.BotAPI, u *tgbotapi.Update)
}

func initBot(c *Config) (*bot, error) {

	botAPI, err := tgbotapi.NewBotAPI(c.BotToken)
	if err != nil {
		return nil, err
	}

	// bot.Debug = true

	m := botDB.NewModel(c.DB)
	dh := newDbHandler(m)
	uh := newTgUpdHandler(m)

	return &bot{
		name: c.BotName,
		api:  botAPI,
		dbh:  dh,
		tgh:  uh,
	}, nil
}

// Start creates the bot instance, checks
// the telegram webhook (if there is none then
// it tries to install it), register appropriate
// handlers on endpoints and start to listen for updates
func Start(c *Config) error {
	bot, err := initBot(c)
	if err != nil {
		return err
	}

	log.Printf("%s: Telegram	->	Authorized on account: [%s]\n", bot.name, bot.api.Self.UserName)

	msg, err := bot.checkWebhook(c)
	if err != nil {
		return err
	}

	log.Printf("%s: Telegram	->	Webhook cheсk: [%s]\n", bot.name, msg)

	updates := bot.api.ListenForWebhook("/" + bot.api.Token)

	http.Handle("/"+c.DbUpdateToken, bot.dbh)

	go http.ListenAndServe(":"+c.Port, nil)

	for update := range updates {
		log.Printf("%s: Telegram	->	Update received: chatID-[%d], user-[%v], text[%s]\n",
			bot.name, update.Message.Chat.ID, update.Message.From, update.Message.Text)

		go bot.tgh.handleUpdate(bot.api, &update)
	}

	return nil
}

func (bot *bot) checkWebhook(c *Config) (string, error) {

	info, err := bot.api.GetWebhookInfo()
	if err != nil {
		return "", err
	}

	if info.LastErrorDate != 0 {
		return "", fmt.Errorf("telegram callback failed: %s", info.LastErrorMessage)
	}

	if info.IsSet() {
		return "webhook already set and available", nil
	}

	_, err = bot.api.SetWebhook(tgbotapi.NewWebhook(c.AppURL + "/" + bot.api.Token))
	if err != nil {
		return "", err
	}

	return "new webhook installed", nil
}
