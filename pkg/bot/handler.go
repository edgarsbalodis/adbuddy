package bot

import (
	"log"
	"strings"

	"github.com/edgarsbalodis/adbuddy/pkg/questionnare"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/exp/slices"
)

// MAIN HANDLER
// processes the user response according to the type of incoming message/callback
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

// ----- TODO: SEPERATE THIS IN SEPERATE FILE -----

var allowedIds = []int64{1268266402, 5948083859}

// checks if user is allowed to communicate with bot
// it can be based on subscription/tokens whatever
func isUserValid(userID int64) bool {
	return slices.Contains(allowedIds, userID)
}

// ----- END -----

var initialQuestion string = "What are you looking for?"
var initialOptions = map[string]string{
	"residence": "Residence",
	"flat":      "Flats",
}

// TODO:
// Based on first response (type of ads)
func returnBaseQuestions(adtype string) questionnare.QuestionnareList {
	var questions questionnare.QuestionnareList
	switch adtype {
	case "residence", "flat":
		questions = questionnare.QuestionnareList{
			&questionnare.Questionnare{
				Question: initialQuestion,
				Options:  initialOptions,
				Key:      "Type",
			},
			&questionnare.Questionnare{
				Question: "Do you want to buy or rent?",
				Options: map[string]string{
					"sell": "Buy",
					"rent": "Rent",
				},
				Key: "TransactionType",
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
			&questionnare.Questionnare{
				Question: "Choose region/city",
				Options: map[string]string{
					"riga-region": "Rīgas raj.",
					"riga":        "Rīga",
					"jurmala":     "Jūrmala",
				},
				Key: "Region",
			},
		}
		return questions
	default:

	}
	return questions
}

func returnSubRegion(region string) *questionnare.Questionnare {
	if strings.Contains(region, "riga-region") {
		return &questionnare.Questionnare{
			Question: "Choose sub region",
			Options: map[string]string{
				"marupes-pag": "Mārupes pag.",
				"babites-pag": "Babītes pag.",
			},
			Key:      "Subregion",
			FollowUp: true,
		}
	} else if strings.Contains(region, "riga") {
		return &questionnare.Questionnare{
			Question: "Choose sub region",
			Options: map[string]string{
				"centre":     "Centre",
				"agenskalns": "Āgenskalns",
			},
			Key:      "Subregion",
			FollowUp: true,
		}
	} else {
		return &questionnare.Questionnare{}
	}
}

// func returnPriceAndRoomQ() questionnare.QuestionnareList {
// 	return questionnare.QuestionnareList{}
// }

func (b *Bot) handleUserResponses(chatID int64, ctx *UserContext, userContexts UserContextMap, text string) {
	// get question's answer
	myKey := ctx.Questionnares[ctx.CurrentQuestion].Key

	// when I get type, then based on it I can get other questions to set up filters
	switch ctx.CurrentQuestion {
	case 0:
		q := returnBaseQuestions(text)
		ctx.Questionnares = q
	case 3:
		q := returnSubRegion(text)
		ctx.Questionnares = append(ctx.Questionnares, q)
	}

	if ctx.Questionnares[ctx.CurrentQuestion].FollowUp {
		// send follow-up question
		if val, ok := ctx.Answers[myKey]; ok {
			if s, ok := val.([]string); ok {
				ctx.Answers[myKey] = append(s, text)
			} else {
				ctx.Answers[myKey] = []string{}
			}
		} else {
			ctx.Answers[myKey] = append([]string{}, text)
		}
		// todo check if there are any options left, if there are not, then ask for next question
		followupQ := "Do you want to add more?"
		buttonYes := tgbotapi.NewInlineKeyboardButtonData("Yes", "follow_up_yes")
		buttonNo := tgbotapi.NewInlineKeyboardButtonData("No", "follow_up_no")
		row := []tgbotapi.InlineKeyboardButton{buttonYes, buttonNo}
		markup := tgbotapi.NewInlineKeyboardMarkup(row)
		msg := tgbotapi.NewMessage(chatID, followupQ)
		msg.ReplyMarkup = markup
		b.tgBot.Send(msg)
		return
	} else {
		ctx.Answers[myKey] = text
	}
	questions := ctx.Questionnares

	// create function - for ask next question
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

		// TODO: make this as function !!!
		// filter := filters.NewEmptyFilter(ctx.Answers["Type"].(string))
		// if filter != nil {
		// 	v := reflect.ValueOf(filter).Elem()
		// 	for key, val := range ctx.Answers {
		// 		if key == "Type" {
		// 			continue
		// 		}
		// 		v.FieldByName(key).SetString(val.(string))
		// 	}
		// }
		// coll := b.client.Database("adbuddy").Collection("responses")
		// response := storage.Response{
		// 	UserID:    ctx.ChatUser.UserID,
		// 	ChatID:    ctx.ChatUser.ChatID,
		// 	Type:      ctx.Answers["Type"].(string),
		// 	Timestamp: time.Now(),
		// 	Answers:   []storage.Answer{},
		// }

		// for k, v := range ctx.Answers {
		// 	if k == "Type" {
		// 		continue
		// 	}
		// 	response.Answers = append(response.Answers, storage.Answer{
		// 		Key:   k,
		// 		Value: v.(string),
		// 	})
		// }

		// // insert in DB
		// _, err := coll.InsertOne(context.TODO(), response)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// ads := scraper.ScrapeData(filter)
		// for _, ad := range ads {
		// 	msg := tgbotapi.NewMessage(chatID, ad.Url)
		// 	b.tgBot.Send(msg)
		// }

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

// contains checks if a string is present in a slice of strings
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
