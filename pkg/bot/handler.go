package bot

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/edgarsbalodis/adbuddy/pkg/questionnare"
	"github.com/edgarsbalodis/adbuddy/pkg/storage"
	"github.com/edgarsbalodis/scraper/pkg/filters"
	"github.com/edgarsbalodis/scraper/pkg/scraper"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/exp/slices"
)

// Survey
type Survey struct {
	Questions []QuestionAnswerPair
}

type QuestionAnswerPair struct {
	Question string
	Answers  map[string]string
	Key      string
}

func (b *Bot) Handler(update tgbotapi.Update, userContexts UserContextMap) {
	switch {
	case update.CallbackQuery != nil:
		b.handleQueryCallback(update, userContexts)
	case update.Message != nil:
		b.handleMessage(update, userContexts)
	default:
		return
	}
}

var allowedIds = []int64{1268266402, 5948083859}

// checks if user is allowed to communicate with bot
// it can be based on subscription/tokens whatever
func isUserValid(userID int64) bool {
	return slices.Contains(allowedIds, userID)
}

var initialQuestion string = "What are you looking for?"
var initialOptions = map[string]string{
	"residence": "Residence",
	"car":       "Car",
}

func (b *Bot) handleQueryCallback(update tgbotapi.Update, userContexts UserContextMap) {
	// there cannot be a case that user is allowed to send callback message if not valid
	if !isUserValid(update.CallbackQuery.From.ID) {
		b.sendNotAllowedMessage(update.CallbackQuery.Message.Chat.ID)
		return
	}

	// check for two cases where string contains use_btn || delete_btn
	// those callbacks come from /filters command
	// TODO: use btn functionality
	// TODO: delete_btn functionality

	answer := update.CallbackQuery.Data
	userID := update.CallbackQuery.From.ID
	chatID := update.CallbackQuery.Message.Chat.ID

	if ctx, exists := userContexts[userID]; exists {
		// checks userContexts and if in-progress with this userID user, then -> handleUserResponses()
		b.handleUserResponses(chatID, ctx, userContexts, answer)
	} else {
		return
	}
}

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
	// b.sendGreetingtMessage(chatID)

	// TODO:
	// GET FIRST QUESTION FROM DB
	button1 := tgbotapi.NewInlineKeyboardButtonData("Residences", "residence")
	button2 := tgbotapi.NewInlineKeyboardButtonData("Cars", "car")

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

// TODO:
// Based on first response (type of ads)
func returnSurveyQuestions(adtype string) questionnare.QuestionnareList {
	var questions questionnare.QuestionnareList
	switch adtype {
	case "residence":
		questions = questionnare.QuestionnareList{
			&questionnare.Questionnare{
				Question: initialQuestion,
				Options:  initialOptions,
				Key:      "Type",
			},
			&questionnare.Questionnare{
				Question: "Choose region",
				Options: map[string]string{
					"riga-region": "Rīgas raj.",
				},
				Key: "Region",
			},
			&questionnare.Questionnare{
				Question: "Choose sub region",
				Options: map[string]string{
					"marupes-pag": "Mārupes pag.",
					"babites-pag": "Babītes pag",
				},
				Key: "Subregion",
			},
			&questionnare.Questionnare{
				Question: "What's your max price? For example: 300000",
				Options:  map[string]string{},
				Key:      "Price",
			},
			&questionnare.Questionnare{
				Question: "Min rooms? For example: 4",
				Options:  map[string]string{},
				Key:      "NumOfRooms",
			},
		}
		return questions
	default:

	}
	return questions
}

func (b *Bot) handleUserResponses(chatID int64, ctx *UserContext, userContexts UserContextMap, text string) {
	// get question's answer
	myKey := ctx.Questionnares[ctx.CurrentQuestion].Key
	ctx.Answers[myKey] = text

	if ctx.CurrentQuestion == 0 {
		questions := returnSurveyQuestions(text)
		ctx.Questionnares = questions
	}
	questions := ctx.Questionnares

	if ctx.CurrentQuestion+1 < len(questions) {
		// increment 1 to get next question
		ctx.CurrentQuestion++

		// get question
		q := questions[ctx.CurrentQuestion]

		if len(q.Options) > 0 {
			row := []tgbotapi.InlineKeyboardButton{}
			for key, val := range q.Options {
				button := tgbotapi.NewInlineKeyboardButtonData(val, key)

				// append to row
				row = append(row, button)
			}
			markup := tgbotapi.NewInlineKeyboardMarkup(row)
			msg := tgbotapi.NewMessage(chatID, q.Question)
			msg.ReplyMarkup = markup

			_, err := b.tgBot.Send(msg)
			if err != nil {
				log.Panic(err)
			}
		} else {
			msg := tgbotapi.NewMessage(chatID, q.Question)
			b.tgBot.Send(msg)
		}
	} else {
		b.sendSuccessMessage(chatID)

		filter := filters.CreateEmptyFilter(ctx.Answers["Type"])
		if filter != nil {
			v := reflect.ValueOf(filter).Elem()
			for key, val := range ctx.Answers {
				if key == "Type" {
					continue
				}
				v.FieldByName(key).SetString(val)
			}
		}
		coll := b.client.Database("adbuddy").Collection("responses")
		response := storage.Response{
			UserID:    ctx.ChatUser.UserID,
			ChatID:    ctx.ChatUser.ChatID,
			Type:      ctx.Answers["Type"],
			Timestamp: time.Now(),
			Answers:   []storage.Answer{},
		}

		for k, v := range ctx.Answers {
			if k == "Type" {
				continue
			}
			response.Answers = append(response.Answers, storage.Answer{
				Key:   k,
				Value: v,
			})
		}

		// insert in DB
		_, err := coll.InsertOne(context.TODO(), response)
		if err != nil {
			log.Fatal(err)
		}

		ads := scraper.ScrapeData(filter)
		for _, ad := range ads {
			msg := tgbotapi.NewMessage(chatID, ad.Url)
			b.tgBot.Send(msg)
		}

		// Delete context to end the survey
		delete(userContexts, ctx.ChatUser.UserID)
	}
	// }
	// fmt.Printf("The answer is: %v", text)
	// append answer to users context
	// ask for survery questions

	// send thank you for answers and wait for results
	// start scraping function
	// delete user's context when done
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
		addButton := tgbotapi.NewInlineKeyboardButtonData("Use", "use_btn_"+val.ID.String())
		delButton := tgbotapi.NewInlineKeyboardButtonData("Delete", "delete_btn_"+val.ID.String())
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
