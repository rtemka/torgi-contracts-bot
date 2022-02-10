package main

import (
	"log"
	"os"
	trbot "trbot/src/bot"
	botDB "trbot/src/botDB"
)

const (
	appURL        string = "https://torgi-contracts-bot.herokuapp.com"
	botToken      string = "2003091653:AAHHuYqtRHcF2HZoHm3wbRUpaMlu2qEnws8"
	dbUpdateToken string = "KMZ4aV0pffnvepuQY3YsGIYghtsy1Thq"
	botName       string = "torgi-contracts-bot"
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

	c := trbot.Config{
		BotName:       botName,
		Port:          port,
		AppURL:        appURL,
		BotToken:      botToken,
		DbUpdateToken: dbUpdateToken,
		DB:            db,
	}

	err = trbot.Start(&c)
	if err != nil {
		log.Fatal(err)
	}
}
