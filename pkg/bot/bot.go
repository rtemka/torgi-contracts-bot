package bot

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	botDB "tbot/pkg/db"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/gorilla/mux"
)

// Config is the settigns
// for the bot instance
type Config struct {
	AllowedChats     map[int64]bool
	NotificationChat int64
	DB               *sql.DB
	BotName          string
	AppURL           string
	BotToken         string
	DbUpdateToken    string
	UptimeToken      string
}

// Bot is API
type Bot struct {
	r      *mux.Router
	logger *log.Logger
	tgh    tgUpdateHandler
	db     db
	dbUpd  chan struct{}
}

func New(c *Config) (*Bot, error) {
	tgapi, err := initTelegramApi(c)
	if err != nil {
		return nil, err
	}

	logger := log.New(os.Stderr, "["+c.BotName+"] | ", log.LstdFlags|log.Lmsgprefix)

	d := botDB.NewBotDB(c.DB)

	bot := Bot{
		r:      mux.NewRouter(),                                   // app mux router
		db:     d,                                                 // database interface
		logger: logger,                                            // app logger
		tgh:    newTgUpdHandler(logger, d, tgapi, c.AllowedChats), // telegram updates handler
		dbUpd:  make(chan struct{}),                               // database update channel
	}

	if c.NotificationChat != 0 {
		var ntf notifier = newTgNotifier(logger, d, tgapi, c.NotificationChat, bot.dbUpd)
		go ntf.notify() // spin off the notifier in it's own routine
	}

	bot.endpoints(c) // set bot endpoints and middleware

	return &bot, nil
}

// db is responsible for the execution
// of CRUD operations over the database
type db interface {
	Upsert(io.ReadCloser) error
	Delete(io.ReadCloser) error
}

// notifier is the logic responsible for
// sending notifications to specific chat
type notifier interface {
	notify()
}

// tgUpdateHandler is the logic responsible for
// processing incoming telegram updates i.e.
// responding to the users messages
type tgUpdateHandler interface {
	handleUpdate(u *tgbotapi.Update)
}

// Router returns Bot router
func (bot *Bot) Router() *mux.Router {
	return bot.r
}

func (bot *Bot) endpoints(c *Config) {
	bot.r.Use(bot.headersMiddleware, bot.closerMiddleware) // middleware

	// endpoint handlers
	bot.r.Handle("/"+c.DbUpdateToken, bot.enforceJsonMiddleware(bot.dbUpdateHandler(3*time.Second))).
		Methods(http.MethodPost, http.MethodOptions)
	bot.r.HandleFunc("/"+c.UptimeToken, bot.uptimeHandler).Methods(http.MethodGet, http.MethodOptions)
	bot.r.HandleFunc("/"+c.BotToken, bot.telegramUpdateHandler).Methods(http.MethodPost, http.MethodOptions)
}

// closerMiddleware drains and close request body at the end
// of every handler work
func (_ *Bot) closerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		_, _ = io.Copy(io.Discard, r.Body)
		_ = r.Body.Close()
	})
}

// headersMiddleware sets headers for all handlers
func (_ *Bot) headersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}

// enforceJsonMiddleware checks Content-Type header
// and aborted further handlers if mime type is unsupported
func (_ *Bot) enforceJsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		mt, _, err := mime.ParseMediaType(ct)
		if err != nil {
			writeResponse(w, "Malformed Content-Type header", http.StatusBadRequest)
			return
		}

		if mt != "application/json" {
			writeResponse(w, "Content-Type is not 'application/json'", http.StatusUnsupportedMediaType)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// writeResponse is helper function that writes response message to w
func writeResponse(w http.ResponseWriter, message string, httpStatusCode int) {
	w.WriteHeader(httpStatusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"response": message,
	})
}

// uptimeHandler recieves uptime call from external resource
// such as https://uptimerobot.com/ to keep application alive,
// because heroku will force app to sleep when it's idling
func (bot *Bot) uptimeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	bot.logger.Printf("[Uptime] -> [uptime checkup success]")
}

// telegramUpdateHandler decodes request body into
// the telegram update struct and answers with
// appropriate message in telegram chat
func (bot *Bot) telegramUpdateHandler(w http.ResponseWriter, r *http.Request) {
	var update tgbotapi.Update
	_ = json.NewDecoder(r.Body).Decode(&update)
	bot.tgh.handleUpdate(&update)
}

func initTelegramApi(c *Config) (*tgbotapi.BotAPI, error) {
	botAPI, err := tgbotapi.NewBotAPI(c.BotToken)
	if err != nil {
		return nil, err
	}

	// botAPI.Debug = true // degug mode

	log.Printf("[%s] | [Telegram] -> [Authorized on account: %s]", c.BotName, botAPI.Self.UserName)

	msg, err := checkWebhook(botAPI, c)
	if err != nil {
		return nil, err
	}
	log.Printf("[%s] | [Telegram] -> [Webhook check: %s]", c.BotName, msg)

	return botAPI, nil
}

func checkWebhook(tgapi *tgbotapi.BotAPI, c *Config) (string, error) {

	info, err := tgapi.GetWebhookInfo()
	if err != nil {
		return "", err
	}

	if info.IsSet() {
		return "webhook already set and available", nil
	}

	_, err = tgapi.SetWebhook(tgbotapi.NewWebhook(c.AppURL + "/" + tgapi.Token))
	if err != nil {
		return "", err
	}

	return "new webhook installed", nil
}
