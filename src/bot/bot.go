package bot

import (
	"database/sql"
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
	Chats         map[int64]bool
	NotifChat     int64
	UptimeToken   string
}

// bot is the main controller
// over the application logic
type bot struct {
	name  string
	api   *tgbotapi.BotAPI
	dbh   dbUpdateHandler
	tgh   tgUpdateHandler
	ntf   notifier
	chats map[int64]bool
	done  chan struct{}
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
	handleUpdate(u *tgbotapi.Update)
}

// notifier is the logic responsible for
// sending notifications to specific chat
type notifier interface {
	notify()
}

func initBot(c *Config) (*bot, error) {

	botAPI, err := tgbotapi.NewBotAPI(c.BotToken)
	if err != nil {
		return nil, err
	}

	// bot.Debug = true

	// create database update channel
	dbUpd := make(chan struct{})
	// create done channel for notifier
	done := make(chan struct{})

	m := botDB.NewModel(c.DB)
	dh := newDbHandler(c.BotName+": [DB Update Handler]\t->\t", m, dbUpd)
	uh := newTgUpdHandler(c.BotName+": [Telegram Update Handler]\t->\t", m, botAPI)
	n := newTgNotifier(c.BotName+": [Notifier]\t->\t", m, botAPI, c.NotifChat, dbUpd, done)

	return &bot{
		name:  c.BotName,
		api:   botAPI,
		dbh:   dh,
		tgh:   uh,
		ntf:   n,
		chats: c.Chats,
		done:  done,
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
	defer func() {
		close(bot.done)
	}()

	log.Printf("%s: [Telegram]\t->\tAuthorized on account: [%s]\n", bot.name, bot.api.Self.UserName)

	msg, err := bot.checkWebhook(c)
	if err != nil {
		return err
	}

	log.Printf("%s: [Telegram]\t->\tWebhook check: [%s]\n", bot.name, msg)

	updates := bot.api.ListenForWebhook("/" + bot.api.Token) // handle telegram webhook update

	http.Handle("/"+c.DbUpdateToken, bot.dbh) // handle DB update

	http.Handle("/"+c.UptimeToken, http.HandlerFunc(bot.uptimeHandler)) // handle uptime check

	go http.ListenAndServe(":"+c.Port, nil)

	// spin off the notifier in it's own routine
	go bot.ntf.notify()

	for update := range updates {

		log.Printf("%s: [Telegram]\t->\tUpdate received: chatID-[%d], user-[%v], text[%s]\n",
			bot.name, update.Message.Chat.ID, update.Message.From, update.Message.Text)

		if !bot.chats[update.Message.Chat.ID] {
			log.Printf("%s: [Telegram]\t->\tchatID-[%d] is not valid... skipped\n", bot.name, update.Message.Chat.ID)
			continue
		}

		go bot.tgh.handleUpdate(&update)
	}

	return nil
}

func (bot *bot) checkWebhook(c *Config) (string, error) {

	info, err := bot.api.GetWebhookInfo()
	if err != nil {
		return "", err
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

// uptimeHandler recieves uptime call from external resource
// such as https://uptimerobot.com/ to keep application alive,
// because heroku will force app to sleep when it's idling
func (b *bot) uptimeHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte{})
	if err != nil {
		log.Printf("%s: [Uptime handler]\t->\tuptime checkup failure [%s]\n", b.name, err.Error())
		return
	}
	log.Printf("%s: [Uptime handler]\t->\tuptime checkup success\n", b.name)
}
