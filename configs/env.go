package configs

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Env struct {
	AppEnv            string
	TimeLocation      *time.Location
	DSN               string
	BotID             int64
	Token             string
	HelpMessageID     int
	HelpMessageChatID int64
	APIEndpoint       string
}

func NewEnv() Env {

	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found or failed to load")
	}

	timeLocation, _ := time.LoadLocation(os.Getenv("TIME_LOCATION"))
	BotID, _ := strconv.ParseInt(os.Getenv("BOT_ID"), 10, 64)
	HelpMessageID, _ := strconv.ParseInt(os.Getenv("HELP_MESSAGE_ID"), 10, 32)
	HelpMessageChatID, _ := strconv.ParseInt(os.Getenv("HELP_MESSAGE_CHAT_ID"), 10, 64)

	env := Env{
		AppEnv:            os.Getenv("APP_ENV"),
		APIEndpoint:       os.Getenv("APIENDPOINT"),
		TimeLocation:      timeLocation,
		DSN:               os.Getenv("DSN"),
		BotID:             BotID,
		Token:             os.Getenv("TOKEN"),
		HelpMessageID:     int(HelpMessageID),
		HelpMessageChatID: HelpMessageChatID,
	}

	if env.AppEnv == "development" {
		log.Println("The App is running in development env")
	}

	return env
}
