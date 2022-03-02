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
	appURL  = "https://torgi-contracts-bot.herokuapp.com"
	botName = "torgi-contracts-bot"
)

// environment variable
var (
	port          string
	dbParams      string
	chats         string
	botToken      string
	dbUpdateToken string
	notifChat     string
)

// getEnvs gets all required environment vars
func getEnvs() error {
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

	return nil
}

func main() {

	if err := getEnvs(); err != nil {
		log.Fatal(err)
	}

	db, err := botDB.OpenDB(dbParams)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	s := strings.Split(chats, " ")
	validChats := make(map[int64]bool, len(s))

	for i := range s {
		n, err := strconv.ParseInt(s[i], 10, 0)
		if err != nil {
			log.Fatal(err)
		}
		validChats[n] = true
	}

	nChat, err := strconv.ParseInt(notifChat, 10, 0)
	if err != nil {
		log.Fatal(err)
	}

	c := trbot.Config{
		BotName:       botName,
		Port:          port,
		AppURL:        appURL,
		BotToken:      botToken,
		DbUpdateToken: dbUpdateToken,
		DB:            db,
		Chats:         validChats,
		NotifChat:     nChat,
	}

	err = trbot.Start(&c)
	if err != nil {
		log.Fatal(err)
	}
}
