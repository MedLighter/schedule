package main

import (
	"context"
	"errors"
	"log"
	"myxxa/scheduler/internal/bot"
	"myxxa/scheduler/internal/db"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	admin, exists := os.LookupEnv("adminId")
	if !exists {
		log.Fatal(errors.New(".env не нашел adminId"))
	}
	ctx := context.WithValue(context.Background(), "adminId", admin)
	instanceTg, err := bot.NewInstance()
	if err != nil {
		log.Fatal(err)
	}
	var cache = db.New(time.Hour*12, time.Hour*12)
	bot.StartBot(ctx, instanceTg, cache)
}
