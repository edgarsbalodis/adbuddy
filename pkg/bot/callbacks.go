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
		value := strings.TrimPrefix(answer, "delete_btn_")
		b.handleDeleteAnswerCallback(chatID, userID, value)
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

// TODO:
// Finish this
// After deleting, maybe I can edit last message "/filters" without deleted answer
func (b *Bot) handleDeleteAnswerCallback(chatID int64, userID int64, responseID string) {
	// get response by responseID
	response := b.Storage.GetResponse(responseID)
	if response.UserID == 0 {
		log.Fatalf("does not exist: %s", response.ID)
	}

	// if userID is not the same cancel
	if response.UserID != userID {
		msg := tgbotapi.NewMessage(chatID, "Couldn't delete this answer")
		b.tgBot.Send(msg)
		return
	}

	b.Storage.DeleteResponse(responseID)

	msg := tgbotapi.NewMessage(chatID, "Answer deleted successfully")
	b.tgBot.Send(msg)
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
