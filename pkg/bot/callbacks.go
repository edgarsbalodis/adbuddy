package bot

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleQueryCallback(update tgbotapi.Update, userContexts UserContextMap) {
	answer := update.CallbackQuery.Data
	userID := update.CallbackQuery.From.ID
	chatID := update.CallbackQuery.Message.Chat.ID
	// there cannot be a case that user is allowed to send callback message if not valid
	if !b.Storage.IsUserValid(userID) {
		b.sendNotAllowedMessage(chatID)
		return
	}

	switch {
	case strings.Contains(answer, "use_btn_"):
		o := strings.TrimPrefix(answer, "use_btn_")
		b.handleUseFilterCallback(o)
	case strings.Contains(answer, "delete_btn_"):
		// TODO: handles delete filter callback
	case strings.Contains(answer, "follow_up_"):
		value := strings.TrimPrefix(answer, "follow_up_")
		if ctx, ok := userContexts[userID]; ok {
			b.processAnswer(ctx, chatID, value, userContexts)
		}
	default:
		if ctx, exists := userContexts[userID]; exists {
			// checks userContexts and if in-progress with this userID user, then -> handleUserResponses()
			b.handleUserResponses(chatID, ctx, userContexts, answer)
		} else {
			return
		}
	}
}

func (b *Bot) handleUseFilterCallback(responseID string) {
	response := b.Storage.GetResponse(responseID)

	answers := AnswersToMap(response.Answers)
	ads := scrape(answers, response.Timestamp.Format("02.01.2006 15:04"), response.ChatID)

	if len(ads) > 0 {
		for _, ad := range ads {
			msg := tgbotapi.NewMessage(response.ChatID, ad)
			b.tgBot.Send(msg)
		}
	} else {
		msg := tgbotapi.NewMessage(response.ChatID, "There is currently nothing new\n")
		b.tgBot.Send(msg)
	}
	err := b.Storage.UpdateTimestamp(responseID)
	if err != nil {
		log.Print(err)
	}
	fmt.Printf("Response: %v", response)
}

// todo: think of better func name
// processFollowUpAnswer
func (b *Bot) processAnswer(ctx *UserContext, chatID int64, answer string, userContexts UserContextMap) {
	if answer == "no" {
		b.askQuestion(ctx, userContexts, chatID, false)
	} else {
		b.askQuestion(ctx, userContexts, chatID, true)
	}
}
