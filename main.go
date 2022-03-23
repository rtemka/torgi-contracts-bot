package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	trbot "trbot/src/bot"
	botDB "trbot/src/botDB"
)

const (
	botName = "torgi-contracts-bot"
)

// environment variable
var (
	appURL        string
	port          string
	dbParams      string
	chats         string
	botToken      string
	dbUpdateToken string
	uptimeToken   string
	notifChat     string
)

// getEnvs gets all required environment vars
func getEnvs() error {
	appURL = os.Getenv("APP_URL")
	if port == "" {
		return fmt.Errorf("$APP_URL must be set")
	}
	port = os.Getenv("PORT")
	if port == "" {
		return fmt.Errorf("$PORT must be set")
	}
	dbParams = os.Getenv("DATABASE_URL")
	if dbParams == "" {
		return fmt.Errorf("$DATABASE_URL must be set")
	}
	chats = os.Getenv("CHATS")
	if chats == "" {
		return fmt.Errorf("$CHATS must be set")
	}
	botToken = os.Getenv("BOT_TOKEN")
	if botToken == "" {
		return fmt.Errorf("$BOT_TOKEN must be set")
	}
	dbUpdateToken = os.Getenv("DB_UPDATE_TOKEN")
	if dbUpdateToken == "" {
		return fmt.Errorf("$DB_UPDATE_TOKEN must be set")
	}
	notifChat = os.Getenv("NOTIF_CHAT")
	if dbUpdateToken == "" {
		return fmt.Errorf("$NOTIF_CHAT must be set")
	}
	uptimeToken = os.Getenv("UPTIME_TOKEN")
	if dbUpdateToken == "" {
		return fmt.Errorf("$UPTIME_TOKEN must be set")
	}

	return nil
}

// parseValidChats returns chats map parsed from
// environment variable. This function expects that provided
// variable is a string with space separated chat id's.
func parseValidChats(chats string) (map[int64]bool, error) {
	s := strings.Split(chats, " ")
	validChats := make(map[int64]bool, len(s))

	for i := range s {
		n, err := parseChat(s[i])
		if err != nil {
			return nil, err
		}
		validChats[n] = true
	}
	return validChats, nil
}

func parseChat(chat string) (int64, error) {
	n, err := strconv.ParseInt(chat, 10, 0)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func main() {

	// get all required environment vars
	if err := getEnvs(); err != nil {
		log.Fatal(err)
	}

	// establish database connection
	db, err := botDB.OpenDB(dbParams)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// parse allowed chats from environment
	validChats, err := parseValidChats(chats)
	if err != nil {
		log.Fatal(err)
	}

	// parse notification chat
	nChat, err := parseChat(notifChat)
	if err != nil {
		log.Fatal(err)
	}

	// set up configuration for the bot
	c := trbot.Config{
		BotName:       botName,
		Port:          port,
		AppURL:        appURL,
		BotToken:      botToken,
		DbUpdateToken: dbUpdateToken,
		DB:            db,
		Chats:         validChats,
		NotifChat:     nChat,
		UptimeToken:   uptimeToken,
	}

	err = trbot.Start(&c)
	if err != nil {
		log.Fatal(err)
	}
}
