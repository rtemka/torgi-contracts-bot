package bot

import (
	"database/sql"
	"io"
	"log"
	"net/http"
	"os"

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

	// database handler
	m := botDB.NewModel(c.DB)

	// database update request handler
	dhLog := log.New(os.Stderr, c.BotName+": [DB Update Handler]\t->\t", log.LstdFlags|log.Lmsgprefix)
	dh := newDbHandler(dhLog, m, dbUpd)
	// telegram update handler
	uhLog := log.New(os.Stderr, c.BotName+": [Telegram Update Handler]\t->\t", log.LstdFlags|log.Lmsgprefix)
	uh := newTgUpdHandler(uhLog, m, botAPI)
	// notifier
	nLog := log.New(os.Stderr, c.BotName+": [Notifier]\t->\t", log.LstdFlags|log.Lmsgprefix)
	n := newTgNotifier(nLog, m, botAPI, c.NotifChat, dbUpd, done)

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
	defer close(bot.done)

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

		if bot.chats[update.Message.Chat.ID] {
			go bot.tgh.handleUpdate(&update)
		} else {
			log.Printf("%s: [Telegram]\t->\tchatID-[%d] is not valid... skipped\n", bot.name, update.Message.Chat.ID)
		}

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
func (b *bot) uptimeHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		_, _ = io.Copy(io.Discard, r.Body)
		_ = r.Body.Close()
	}()

	w.WriteHeader(http.StatusOK)

	log.Printf("%s: [Uptime handler]\t->\tuptime checkup success\n", b.name)
}
