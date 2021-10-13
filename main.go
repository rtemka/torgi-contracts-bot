package main

import (
	"log"
	"os"
	trbot "trbot/src/bot"
)

const (
	APP_URL string = ""
)

func main() {

	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	c := trbot.Config{
		Port:   port,
		AppURL: APP_URL,
	}

	err := trbot.Start(&c)
	if err != nil {
		log.Fatal(err)
	}
}
