package bot

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/edgarsbalodis/adbuddy/pkg/storage"
	"github.com/edgarsbalodis/scraper/pkg/filters"
	"github.com/edgarsbalodis/scraper/pkg/scraper"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (b *Bot) handleQueryCallback(update tgbotapi.Update, userContexts UserContextMap) {
	// there cannot be a case that user is allowed to send callback message if not valid
	if !isUserValid(update.CallbackQuery.From.ID) {
		b.sendNotAllowedMessage(update.CallbackQuery.Message.Chat.ID)
		return
	}
	answer := update.CallbackQuery.Data
	userID := update.CallbackQuery.From.ID
	chatID := update.CallbackQuery.Message.Chat.ID
	switch {
	case strings.Contains(answer, "use_btn_"):
		// handles use filter callback
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
	id, err := primitive.ObjectIDFromHex(responseID)
	if err != nil {
		log.Println("Invalid id")
	}
	coll := b.client.Database("adbuddy").Collection("responses")
	result := coll.FindOne(context.Background(), bson.M{"_id": id})

	response := storage.Response{}
	result.Decode(&response)
	filter := filters.NewEmptyFilter(response.Type)
	if filter != nil {
		v := reflect.ValueOf(filter).Elem()
		for _, val := range response.Answers {
			v.FieldByName(val.Key).SetString(val.Value)
		}
		v.FieldByName("Timestamp").SetString(response.Timestamp.Format("02.01.2006 15:04"))
	}
	ads := scraper.ScrapeData(filter)
	if len(ads) > 0 {
		for _, ad := range ads {
			msg := tgbotapi.NewMessage(response.ChatID, ad.Url)
			b.tgBot.Send(msg)
		}
	} else {
		msg := tgbotapi.NewMessage(response.ChatID, "There is currently nothing new\n")
		b.tgBot.Send(msg)
	}
	// update timestamp for filter
	f := bson.D{primitive.E{Key: "_id", Value: id}}
	update := bson.D{{"$set", bson.D{{"timestamp", time.Now()}}}}

	res, updateErr := coll.UpdateOne(context.TODO(), f, update)
	if updateErr != nil {
		panic(updateErr)
	}
	if res.ModifiedCount == 0 {
		// Handle the case where no documents were updated
		fmt.Print("update")
	}
	fmt.Printf("Response: %v", response)
}

// todo: think of better func name
func (b *Bot) processAnswer(ctx *UserContext, chatID int64, answer string, userContexts UserContextMap) {
	questions := ctx.Questionnares

	// ask next question
	if answer == "no" {
		ctx.CurrentQuestion++
		if ctx.CurrentQuestion+1 < len(questions) {
			// increment 1 to get next question

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
			// TODO: use function for ending questioning
			b.sendSuccessMessage(chatID)
			delete(userContexts, ctx.ChatUser.UserID)
		}
	} else {
		// get question
		q := questions[ctx.CurrentQuestion]
		// opt := q.Options
		// answer := ctx.Answers[]
		if len(q.Options) > 0 {
			filteredMap := make(map[string]string)
			myKey := ctx.Questionnares[ctx.CurrentQuestion].Key
			for k, val := range q.Options {
				if !contains(ctx.Answers[myKey].([]string), k) {
					filteredMap[k] = val
				}
			}
			row := []tgbotapi.InlineKeyboardButton{}
			for key, val := range filteredMap {
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
	}
}
