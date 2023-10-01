package main

import (
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/exp/slices"
)

var allowedIds = []int{1268266402, 5948083859}

type UserContext struct {
	UserID          int
	CurrentQuestion int
	Questions       []string
	UserAnswers     []string
}

func main1() {
	bot, err := tgbotapi.NewBotAPI("")
	if err != nil {
		log.Panic(err)
	}

	// Initialize a map to store user contexts
	userContexts := make(map[int]*UserContext)

	bot.Debug = true
	// log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message

			userID := update.Message.From.ID

			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			// check if for known user
			if !slices.Contains(allowedIds, int(update.Message.From.ID)) {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Father told me that I can't talk to strangers")
				bot.Send(msg)
				continue
			}

			text := strings.ToLower(update.Message.Text)

			// Check if a conversation is already in progress for this user
			if ctx, exists := userContexts[int(userID)]; exists {
				processUserResponse(bot, update.Message.Chat.ID, ctx, userContexts, text)
			} else {
				processInitialCommands(bot, update.Message.Chat.ID, userContexts, int(userID), text)
			}

			// options := []string{"New cars", "Used cars", "Car rentals"}

			// msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			// msg.ReplyToMessageID = update.Message.MessageID
			// sendOptions(bot, update.Message.Chat.ID, options)

			// bot.Send(msg)
		}
	}
}

func processInitialCommands(bot *tgbotapi.BotAPI, chatID int64, userContexts map[int]*UserContext, userID int, text string) {
	switch text {
	case "/start":
		// Start a conversation with the first question
		initialQuestion := "Where do you want to but a house?"
		questions := []string{
			initialQuestion,
			"What is your favorite animal?",
			"What is your favorite food?",
			"Tell me more about yourself (type your answer).",
		}

		ctx := &UserContext{
			UserID:          userID,
			CurrentQuestion: 0,
			Questions:       questions,
		}

		userContexts[userID] = ctx

		// Send the first question
		msg := tgbotapi.NewMessage(chatID, initialQuestion)
		bot.Send(msg)

	default:
		reply := "I don't understand that command."
		msg := tgbotapi.NewMessage(chatID, reply)
		bot.Send(msg)
	}
}

func processUserResponse(bot *tgbotapi.BotAPI, chatID int64, ctx *UserContext, userContexts map[int]*UserContext, text string) {
	// Record the user's answer
	ctx.UserAnswers = append(ctx.UserAnswers, text)

	// Check if there are more questions
	if ctx.CurrentQuestion+1 < len(ctx.Questions) {
		// Send the next question
		ctx.CurrentQuestion++
		nextQuestion := ctx.Questions[ctx.CurrentQuestion]
		msg := tgbotapi.NewMessage(chatID, nextQuestion)
		bot.Send(msg)
	} else {
		// End the conversation and display the recorded answers
		reply := "Thank you for answering all the questions:\n"
		for i, answer := range ctx.UserAnswers {
			reply += ctx.Questions[i] + ": " + answer + "\n"
		}
		msg := tgbotapi.NewMessage(chatID, reply)
		bot.Send(msg)

		// Remove the user context to end the conversation
		delete(userContexts, ctx.UserID)
	}
}

// func processInitialCommands(bot *tgbotapi.BotAPI, chatID int64, userContexts map[int]*UserContext, userID int, text string) {
// 	switch text {
// 	case "/start":
// 		reply := "Welcome, I will help you find ads 🏡🏎️  Just type /find"
// 		msg := tgbotapi.NewMessage(chatID, reply)
// 		bot.Send(msg)
// 	case "/help":
// 		reply := "Available commands:\n\n"
// 		reply += "/start - Start using the bot\n"
// 		reply += "/help - Show this help message\n"
// 		reply += "/find - Start searching for ads\n"
// 		msg := tgbotapi.NewMessage(chatID, reply)
// 		bot.Send(msg)
// 	case "/find":
// 		sendQuestion(bot, chatID, userAnswers, int64(userID), 0)
// 	default:
// 		reply := "I don't know this command."
// 		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
// 		bot.Send(msg)
// 	}
// }

// func sendOptions(bot *tgbotapi.BotAPI, chatID int64, options []string) {
// 	var buttons []tgbotapi.InlineKeyboardButton

// 	for _, option := range options {
// 		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(option, option))
// 	}

// 	keyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(buttons...))
// 	msg := tgbotapi.NewMessage(chatID, "Please choose one of the following:")
// 	msg.ReplyMarkup = keyboard
// 	bot.Send(msg)
// }

// func sendQuestion(bot *tgbotapi.BotAPI, chatID int64, userAnswers map[int][]string, userID int64, questionIndex int) {
// 	questions := []string{
// 		"Where?",
// 		"What's the price range? (for example: >=300000)",
// 		"Room count (for example: <=4)",
// 	}

// 	// Check if there are more questions to ask
// 	if questionIndex < len(questions) {
// 		// Get the current question
// 		question := questions[questionIndex]

// 		// Create a message to send the question
// 		msg := tgbotapi.NewMessage(chatID, question)

// 		// Check if the current question allows a custom text answer
// 		if questionIndex != 0 {
// 			// If it's not the first question, instruct the user to type their answer
// 			msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}
// 		} else {
// 			// If it's first question, create an inline keyboard with options
// 			currentOptions := []string{"marupes-pag", "babites-pag"} // Replace with options for this question
// 			var replyMarkup tgbotapi.InlineKeyboardMarkup
// 			var keyboardRows []tgbotapi.InlineKeyboardButton

// 			for _, option := range currentOptions {
// 				button := tgbotapi.NewInlineKeyboardButtonData(option, option)
// 				keyboardRows = append(keyboardRows, button)
// 			}

// 			replyMarkup.InlineKeyboard = append(replyMarkup.InlineKeyboard, keyboardRows)
// 			msg.ReplyMarkup = replyMarkup
// 		}

// 		// Send the question to the user
// 		bot.Send(msg)

// 		// Wait for the user's response
// 		update := <-bot.Updates
// 		if update.Message != nil {
// 			answer := update.Message.Text

// 			// Save the user's answer in memory
// 			userAnswers[userID] = append(userAnswers[userID], answer)

// 			// Send a confirmation message
// 			reply := "Your answer has been recorded: " + answer
// 			msg := tgbotapi.NewMessage(chatID, reply)
// 			bot.Send(msg)

// 			// Continue to the next question in the chain
// 			sendQuestion(bot, chatID, userAnswers, userID, questionIndex+1)
// 		}
// 	} else {
// 		// If there are no more questions, end the conversation
// 		reply := "Thank you for answering all the questions!"
// 		msg := tgbotapi.NewMessage(chatID, reply)
// 		bot.Send(msg)
// 	}

// }
