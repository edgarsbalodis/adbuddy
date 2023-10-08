package main

import (
	"context"
	"log"

	"github.com/edgarsbalodis/adbuddy/internal/config"
	"github.com/edgarsbalodis/adbuddy/pkg/bot"
	"github.com/edgarsbalodis/adbuddy/pkg/storage"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// mongo db connection
	client, _ := storage.NewMongoClient(cfg.DatabaseUrl)
	defer client.Disconnect(context.Background())
	storage := storage.New(client)

	// mongodb seeder
	storage.Seed()

	// Create a new bot instance
	botInstance, err := bot.NewBotApi(cfg.TelegramBotToken)
	if err != nil {
		log.Fatalf("Bot encountered an error: %v", err)
	}

	// telegram bot struct
	tgBot := bot.New(botInstance, storage)
	tgBot.StartScraperCron()

	tgBot.StartBot()
}
