package bot

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/edgarsbalodis/adbuddy/pkg/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (b *Bot) handleMessage(update tgbotapi.Update, userContexts UserContextMap) {
	text := strings.ToLower(update.Message.Text)
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	// check if /start command's conversation is in-progress already
	if ctx, exists := userContexts[userID]; exists {
		// there cannot be a case that user is having existing conversation and not valid
		b.handleUserResponses(chatID, ctx, userContexts, text)
	} else {
		switch text {
		case "/start":
			// check for valid userIDs
			if !isUserValid(userID) {
				b.sendNotAllowedMessage(chatID)
				return
			}
			b.handleStartCommand(chatID, userID, userContexts, text)
		case "/filters":
			if !isUserValid(userID) {
				b.sendNotAllowedMessage(chatID)
				return
			}
			b.handleFiltersFunction(userID, chatID)
		case "/help":
			b.handleHelpCommand(chatID)
		default:
			if !isUserValid(userID) {
				b.sendNotAllowedMessage(chatID)
				return
			}
			b.sendUnknownCommandMessage(chatID)
		}
	}
}

func (b *Bot) handleStartCommand(chatID, userID int64, userContexts UserContextMap, text string) {
	q := b.Storage.GetMainQuestion()
	opts := b.Storage.GetFilters()

	row := []tgbotapi.InlineKeyboardButton{}
	var btn tgbotapi.InlineKeyboardButton
	for _, opt := range opts {
		btn = tgbotapi.NewInlineKeyboardButtonData(opt.Name, opt.Value)
		row = append(row, btn)
	}

	chatUser := CreateChatUser(userID, chatID)

	ctx := NewUserContext(chatUser, 0, []storage.Question{q})

	// add this UserContext to userContexts map
	userContexts[userID] = ctx

	// create the keyboard markup with the row of buttons
	markup := tgbotapi.NewInlineKeyboardMarkup(row)

	// ask initial question with provided answers
	msg := tgbotapi.NewMessage(chatID, q.Text)
	msg.ReplyMarkup = markup

	_, err := b.tgBot.Send(msg)
	if err != nil {
		log.Panic(err)
	}
}

func (b *Bot) handleFiltersFunction(userID, chatID int64) {
	// find all responses from db based on chatID and userID
	coll := b.Storage.Client.Database("adbuddy").Collection("responses")
	filter := bson.M{
		"userID": userID,
		"chatID": chatID,
	}
	result, err := coll.Find(context.TODO(), filter)
	if err != nil {
		log.Printf("Error while finding responses: %v", err)
	}
	var data []storage.Response
	if err := result.All(context.Background(), &data); err != nil {
		log.Fatal(err)
	}

	// send responses to chat so user can choose and start scraper with filter immediately
	// Create a keyboard with a button
	for _, val := range data {
		msgText := val.Type + "\n"
		// row := []tgbotapi.InlineKeyboardButton{}
		for _, innerVal := range val.Answers {
			switch v := innerVal.Value.(type) {
			case string:
				msgText += innerVal.Key + ": " + v + "\n"
			case primitive.A:
				var txt string
				for _, item := range v {
					if strVal, ok := item.(string); ok {
						txt += innerVal.Key + ": " + strVal + "\n"
					} else {
						// Handle case where item is not a string
						fmt.Printf("Unexpected type inside array: %T\n", item)
					}
				}
				msgText += txt
			default:
				fmt.Printf("Type of val.Value: %T\n", innerVal.Value)
			}
		}

		addButton := tgbotapi.NewInlineKeyboardButtonData("Use", "use_btn_"+val.ID.Hex())
		delButton := tgbotapi.NewInlineKeyboardButtonData("Delete", "delete_btn_"+val.ID.Hex())
		row := []tgbotapi.InlineKeyboardButton{addButton, delButton}
		markup := tgbotapi.NewInlineKeyboardMarkup(row)
		msg := tgbotapi.NewMessage(chatID, msgText)
		msg.ReplyMarkup = markup

		// Attach the keyboard to the message
		msg.ReplyMarkup = markup

		// Send the message with the button
		_, err = b.tgBot.Send(msg)
		if err != nil {
			log.Panic(err)
		}

	}
}

func (b *Bot) sendFollowUpQuestion(chatID int64) {
	followupQ := "Do you want to add more?"
	buttonYes := tgbotapi.NewInlineKeyboardButtonData("Yes", "follow_up_yes")
	buttonNo := tgbotapi.NewInlineKeyboardButtonData("No", "follow_up_no")
	row := []tgbotapi.InlineKeyboardButton{buttonYes, buttonNo}

	markup := tgbotapi.NewInlineKeyboardMarkup(row)
	msg := tgbotapi.NewMessage(chatID, followupQ)
	msg.ReplyMarkup = markup
	b.tgBot.Send(msg)
}

func (b *Bot) sendSuccessMessage(chatID int64) {
	reply := "Thank you for answering all the questions:\n"
	reply += "Right now I'm searching for ads, as soon as there is something for you I will send it to you\n"

	msg := tgbotapi.NewMessage(chatID, reply)
	_, err := b.tgBot.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (b *Bot) handleHelpCommand(chatID int64) {
	reply := "Welcome to AdBuddyBot, I will help you find ads 🏡 🏎️ \n\n"
	reply += "Available commands:\n"
	reply += "/start - starts conversation about what ads you are looking for\n"
	reply += "/filters - returns previous filters so you can immediately start searching\n"
	reply += "/help - this is message you see right now"

	msg := tgbotapi.NewMessage(chatID, reply)
	_, err := b.tgBot.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (b *Bot) sendUnknownCommandMessage(chatID int64) {
	reply := "I don't know this command"

	msg := tgbotapi.NewMessage(chatID, reply)
	_, err := b.tgBot.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (b *Bot) sendNotAllowedMessage(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Father told me that I can't talk to strangers")
	_, err := b.tgBot.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}
