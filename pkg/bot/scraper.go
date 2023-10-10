package bot

import (
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"

	"github.com/edgarsbalodis/adbuddy/pkg/scraper"
	"github.com/edgarsbalodis/adbuddy/pkg/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AnswersToMap(answers []storage.Answer) AnswersMap {
	ansMap := make(map[string]interface{})
	var subregions []string
	for _, ans := range answers {
		if aval, ok := ans.Value.(primitive.A); ok {
			for _, v := range aval {
				subregions = append(subregions, v.(string))
			}
			ansMap[ans.Key] = subregions
		} else {
			ansMap[ans.Key] = ans.Value
		}
		fmt.Printf("ans %v type: %T", ans.Key, ans.Value)
	}
	return ansMap
}

func scrape(answers AnswersMap, timestamp string, chatID int64) []string {
	var urls []string

	if subregions, ok := answers["Subregion"].([]string); ok {
		resultsChan := make(chan []string)
		var wg sync.WaitGroup

		for _, sr := range subregions {
			filter := scraper.NewEmptyFilter(answers["Type"].(string))
			if filter != nil {
				v := reflect.ValueOf(filter).Elem()
				for key, val := range answers {
					if key == "Subregion" {
						v.FieldByName(key).SetString(sr)
					} else if sVal, ok := val.(string); ok {
						v.FieldByName(key).SetString(sVal)
					}
				}
				v.FieldByName("Timestamp").SetString(timestamp)

				wg.Add(1)
				go func(filter interface{}) {
					defer wg.Done()
					resultsChan <- scraper.ScrapeDataNew(filter)
				}(filter)
			}
		}

		go func() {
			wg.Wait()
			close(resultsChan)
		}()

		for res := range resultsChan {
			urls = append(urls, res...)
		}

	} else {
		filter := scraper.NewEmptyFilter(answers["Type"].(string))
		if filter != nil {
			v := reflect.ValueOf(filter).Elem()
			for key, val := range answers {
				if sVal, ok := val.(string); ok {
					v.FieldByName(key).SetString(sVal)
				}
			}
			v.FieldByName("Timestamp").SetString(timestamp)
			urls = scraper.ScrapeDataNew(filter)
		}
	}

	return urls
}

func (b *Bot) scraperCron() {
	responses := b.Storage.GetResponses()
	if len(responses) < 1 {
		return
	}
	loc, _ := time.LoadLocation("Europe/Riga")
	fmt.Printf("Cron started: %v", time.Now().In(loc))

	for _, r := range responses {
		answerMap := AnswersToMap(r.Answers)
		timestamp := r.Timestamp.In(loc).Format("02.01.2006 15:04")
		ads := scrape(answerMap, timestamp, r.ChatID)

		if len(ads) > 0 {
			for _, ad := range ads {
				msg := tgbotapi.NewMessage(r.ChatID, ad)
				b.tgBot.Send(msg)
			}
		} else {
			msg := tgbotapi.NewMessage(r.ChatID, "There is currently nothing new\n")
			b.tgBot.Send(msg)
		}
		err := b.Storage.UpdateTimestamp(r.ID.Hex())
		if err != nil {
			log.Print(err)
		}
	}
}

func (b *Bot) StartScraperCron() {
	c := cron.New()
	c.AddFunc("@every 1h", func() {
		loc, _ := time.LoadLocation("Europe/Riga")
		t := time.Now().In(loc)
		if t.Hour() > 5 || t.Hour() < 23 {
			b.scraperCron()
		} else {
			return
		}
	})
	c.Start()
}
