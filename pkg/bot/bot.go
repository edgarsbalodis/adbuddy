package bot

import (
	"log"

	"github.com/edgarsbalodis/adbuddy/pkg/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	tgBot   *tgbotapi.BotAPI
	Storage *storage.Storage
}

func New(bot *tgbotapi.BotAPI, storage *storage.Storage) *Bot {
	return &Bot{
		tgBot:   bot,
		Storage: storage,
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
