package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramBotToken string
	DatabaseUrl      string
}

func LoadConfig() *Config {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	telegramBotToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	databaseUrl := os.Getenv("DATABASE_URL")

	return &Config{
		TelegramBotToken: telegramBotToken,
		DatabaseUrl:      databaseUrl,
	}
}
