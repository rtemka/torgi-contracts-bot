package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	trbot "trbot/src/bot"
	botDB "trbot/src/botDB"
)

const (
	appURL        = "https://torgi-contracts-bot.herokuapp.com"
	botToken      = "2003091653:AAHHuYqtRHcF2HZoHm3wbRUpaMlu2qEnws8"
	dbUpdateToken = "KMZ4aV0pffnvepuQY3YsGIYghtsy1Thq"
	botName       = "torgi-contracts-bot"
	chats         = "-1001528623117 113802948 -562535649"
)

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}
	dbParams := os.Getenv("DATABASE_URL")
	if port == "" {
		log.Fatal("$DATABASE_URL must be set")
	}

	db, err := botDB.OpenDB(dbParams)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	s := strings.Split(chats, " ")
	chats := make(map[int64]bool, len(s))

	for i := range s {
		n, err := strconv.ParseInt(s[i], 10, 0)
		if err != nil {
			log.Fatal(err)
		}
		chats[n] = true
	}

	c := trbot.Config{
		BotName:       botName,
		Port:          port,
		AppURL:        appURL,
		BotToken:      botToken,
		DbUpdateToken: dbUpdateToken,
		DB:            db,
		Chats:         chats,
	}

	err = trbot.Start(&c)
	if err != nil {
		log.Fatal(err)
	}
}
