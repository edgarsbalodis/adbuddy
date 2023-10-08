package bot

import (
	"fmt"
	"log"
	"time"

	"github.com/edgarsbalodis/adbuddy/pkg/storage"
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

func (b *Bot) handleUserResponses(chatID int64, ctx *UserContext, userContexts UserContextMap, text string) {
	// get question's answer
	qkey := ctx.Questions[ctx.CurrentQuestion].Key

	// check if user context has only 1 question which is main question
	if len(ctx.Questions) == 1 {
		// save answer
		ctx.Answers[qkey] = text
		// get all questions based on answer
		fq := b.Storage.GetFiltersQuestions(text)
		// append those questions to ctx.Questions
		ctx.Questions = append(ctx.Questions, fq...)
	} else if ctx.Questions[ctx.CurrentQuestion].FollowUp {
		if val, ok := ctx.Answers[qkey]; ok {
			if s, ok := val.([]string); ok {
				ctx.Answers[qkey] = append(s, text)
			} else {
				ctx.Answers[qkey] = []string{}
			}
		} else {
			ctx.Answers[qkey] = append([]string{}, text)
		}

		if isOptionLeft(ctx, qkey) {
			b.sendFollowUpQuestion(chatID)
			return
		} else {
			b.askQuestion(ctx, userContexts, chatID, false)
			return
		}
	} else {
		ctx.Answers[qkey] = text
	}

	b.askQuestion(ctx, userContexts, chatID, false)
}

func isOptionLeft(ctx *UserContext, key string) bool {
	// get answer slice with key
	a := ctx.Answers[key]
	// Check and type assert the value to a slice of strings
	var condition string
	if slice, ok := a.([]string); ok && len(slice) > 0 {
		value := slice[0]
		options := ctx.Questions[ctx.CurrentQuestion].Options
		for _, o := range options {
			if o.Value == value {
				condition = o.Condition
				break
			}
		}

	}
	// i got now ['marupes-pag', 'babites-pag']
	for _, opt := range ctx.Questions[ctx.CurrentQuestion].Options {
		// I need also region value, so I can check versus condition

		// if opt.Value is not in []slice return true else false
		if !slices.Contains(a.([]string), opt.Value) && condition == opt.Condition {
			return true
		}
	}
	return false
}

func (b *Bot) askQuestion(ctx *UserContext, userContexts UserContextMap, chatID int64, followUp bool) {
	var cq int
	if !followUp {
		cq = ctx.CurrentQuestion + 1
	} else {
		cq = ctx.CurrentQuestion
	}
	if cq < len(ctx.Questions) {
		if !followUp {
			ctx.CurrentQuestion++
		}

		q := ctx.Questions[ctx.CurrentQuestion]

		if len(q.Options) > 0 {
			row := []tgbotapi.InlineKeyboardButton{}
			for _, opt := range q.Options {
				// check for condition in options
				if len(opt.Condition) != 0 {
					// condition "riga-region"
					// that has to be value for one of the questions
					// check for value in ctx.Answers
					key := findKeyByValue(ctx.Answers, opt.Condition)
					if key == "" {
						fmt.Print("key not found")
						continue
					}
					if followUp {
						// i need to check answers of this question
						a := ctx.Answers[q.Key]
						// check if this value is in slice
						if slices.Contains(a.([]string), opt.Value) {
							continue
						}
						// if not then add else continue
					}
					button := tgbotapi.NewInlineKeyboardButtonData(opt.Text, opt.Value)
					row = append(row, button)

				} else {
					button := tgbotapi.NewInlineKeyboardButtonData(opt.Text, opt.Value)
					// append to row
					row = append(row, button)
				}
			}
			if len(row) == 0 {
				// if region was 'jurmala'
				// then subregion does not exist, so row == nil and ask next question
				b.askQuestion(ctx, userContexts, chatID, false)
				return
				// there is potential in question mixup,
				// 	basically I ask child question (sub-region)
				// 	even I have not asked parent question (region) yet
				// 	if I update seeder and I add parent_question_id for sub-region question
				//	I can check if this is question with parent, If it is ask parent question or just cache question and when parent question is answered ask child question
			}
			markup := tgbotapi.NewInlineKeyboardMarkup(row)
			msg := tgbotapi.NewMessage(chatID, q.Text)
			msg.ReplyMarkup = markup

			_, err := b.tgBot.Send(msg)
			if err != nil {
				log.Panic(err)
			}
		} else {
			// sends message without multiple choice
			msg := tgbotapi.NewMessage(chatID, q.Text)
			b.tgBot.Send(msg)
		}
	} else {
		b.saveResponse(ctx)
		b.sendSuccessMessage(chatID)

		ads := scrape(ctx.Answers, "", chatID)

		for _, ad := range ads {
			msg := tgbotapi.NewMessage(chatID, ad)
			b.tgBot.Send(msg)
		}
		// Delete context to end the survey
		delete(userContexts, ctx.ChatUser.UserID)
	}
}

func (b *Bot) saveResponse(ctx *UserContext) {
	loc, err := time.LoadLocation("Europe/Riga")
	if err != nil {
		fmt.Print(err)
	}

	response := storage.Response{
		UserID:    ctx.ChatUser.UserID,
		ChatID:    ctx.ChatUser.ChatID,
		Type:      ctx.Answers["Type"].(string),
		Timestamp: time.Now().In(loc),
		Answers:   []storage.Answer{},
	}

	for k, v := range ctx.Answers {
		// Check the actual type of v
		switch val := v.(type) {
		case string:
			response.Answers = append(response.Answers, storage.Answer{
				Key:   k,
				Value: val,
			})
		case []string:
			response.Answers = append(response.Answers, storage.Answer{
				Key:   k,
				Value: val,
			})
		default:
			// Handle unexpected type or ignore
			fmt.Printf("Unexpected type %T for key %s\n", v, k)
		}

	}
	b.Storage.SaveResponse(response)
}

func findKeyByValue(data AnswersMap, targetValue string) string {
	for key, value := range data {
		if strValue, ok := value.(string); ok && strValue == targetValue {
			return key
		}
	}
	return ""
}

// contains checks if a string is present in a slice of strings
// func contains(slice []string, str string) bool {
// 	for _, s := range slice {
// 		if s == str {
// 			return true
// 		}
// 	}
// 	return false
// }
