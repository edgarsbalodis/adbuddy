package bot

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/edgarsbalodis/adbuddy/pkg/questionnare"
	"github.com/edgarsbalodis/adbuddy/pkg/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson"
)

func (b *Bot) handleMessage(update tgbotapi.Update, userContexts UserContextMap) {
	text := strings.ToLower(update.Message.Text)
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID
	// // check if /start command's conversation is in-progress already
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
			b.sendUnknownCommandMessage(chatID)
		}
	}
}

func (b *Bot) handleStartCommand(chatID, userID int64, userContexts UserContextMap, text string) {
	// TODO:
	// GET FIRST QUESTION FROM DB
	button1 := tgbotapi.NewInlineKeyboardButtonData("Residences", "residence")
	button2 := tgbotapi.NewInlineKeyboardButtonData("Flats", "flat")

	chatUser := CreateChatUser(userID, chatID)

	q := questionnare.NewQuestionnare(initialQuestion, initialOptions, "Type")
	ql := questionnare.QuestionnareList{q}

	ctx := NewUserContext(chatUser, 0, ql)

	// add this UserContext to userContexts map
	userContexts[userID] = ctx

	// create row of buttons
	row := []tgbotapi.InlineKeyboardButton{button1, button2}

	// create the keyboard markup with the row of buttons
	markup := tgbotapi.NewInlineKeyboardMarkup(row)

	// ask initial question with provided answers
	msg := tgbotapi.NewMessage(chatID, initialQuestion)
	msg.ReplyMarkup = markup

	_, err := b.tgBot.Send(msg)
	if err != nil {
		log.Panic(err)
	}
}

func (b *Bot) handleFiltersFunction(userID, chatID int64) {
	// find all responses from db based on chatID and userID
	coll := b.client.Database("adbuddy").Collection("responses")
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
	fmt.Print(data)

	// send responses to chat so user can choose and start scraper with filter immediately
	// Create a keyboard with a button
	for _, val := range data {
		msgText := val.Type + "\n"
		// row := []tgbotapi.InlineKeyboardButton{}
		for _, val := range val.Answers {
			msgText += val.Key + ": " + val.Value + "\n"

			// append to row
			// row = append(row, button)
		}
		// objectID := val.ID.Hex()
		// if !ok {
		// 	log.Fatal("Failed to retrieve ObjectId from document")
		// }

		// // Convert the ObjectId to a string
		// objectIDStr := objectID.Hex()
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
