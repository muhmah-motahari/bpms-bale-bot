package configs

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Env struct {
	AppEnv         string
	BaleAPIAddress string
	TimeLocation   *time.Location
	DSN            string
	BotID          int64
	Token          string
}

func NewEnv() Env {
	timeLocation, _ := time.LoadLocation(os.Getenv("TIME_LOCATION"))
	BotID, _ := strconv.ParseInt(os.Getenv("BOT_ID"), 10, 64)

	env := Env{
		AppEnv:         os.Getenv("APP_ENV"),
		BaleAPIAddress: "https://tapi.bale.ai/bot" + os.Getenv("BALE_TOKEN") + "/",
		TimeLocation:   timeLocation,
		DSN:            os.Getenv("DSN"),
		BotID:          BotID,
		Token:          os.Getenv("BALE_TOKEN"),
	}

	if env.AppEnv == "development" {
		log.Println("The App is running in development env")
	}

	return env
}
