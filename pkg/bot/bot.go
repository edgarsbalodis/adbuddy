package bot

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/mongo"
)

type Bot struct {
	client *mongo.Client
	tgBot  *tgbotapi.BotAPI
}

func NewBot(client *mongo.Client, bot *tgbotapi.BotAPI) *Bot {
	return &Bot{
		client: client,
		tgBot:  bot,
	}
}

func NewBotApi(token string) (*tgbotapi.BotAPI, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Printf("Error creating bot: %v", err)
		return nil, err
	}

	// bot setting
	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)
	return bot, nil
}

func (b *Bot) StartBot() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.tgBot.GetUpdatesChan(u)
	userContexts := make(map[int64]*UserContext)

	for update := range updates {
		b.Handler(update, userContexts)
	}
}
