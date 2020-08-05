package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/line/line-bot-sdk-go/linebot"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

type PostRequest struct {
	PostContent string `json:"post"`
}

type Schedule struct {
	Title map[string]*string
	Start map[string]*string
	End   map[string]*string
}

const (
	titleKey = "0"
	startKey = "1"
	endKey   = "2"
)

func DefaultMessage(bot *linebot.Client, event *linebot.Event) error {
	reply := linebot.NewTextMessage("今日も志を忘れず頑張ってください！！")
	if _, err := bot.ReplyMessage(event.ReplyToken, reply).Do(); err != nil {
		log.Print(err)
		return err
	}
	return nil
}

func DatetimeAction(SorE string, bot *linebot.Client, event *linebot.Event) error {
	var timing string
	if SorE == startKey {
		timing = "start"
	} else if SorE == endKey {
		timing = "end"
	}
	reply := linebot.NewTemplateMessage(
		"this is a button template",
		linebot.NewButtonsTemplate(
			"https://shunsuarez.com/calendar.jpg",
			"Hi! I'm Calen",
			"When your schedule is "+timing,
			linebot.NewDatetimePickerAction("Make an appointment", SorE, "datetime", "", "", ""),
		),
	)
	_, err := bot.ReplyMessage(event.ReplyToken, reply).Do()
	if err != nil {
		log.Print(err)
		return err
	}
	return nil
}

func DatetimePB(bot *linebot.Client, event *linebot.Event, sche *Schedule) error {
	if event.Postback.Data == startKey {
		sche.Start[startKey] = &event.Postback.Params.Datetime
		if err := DatetimeAction(endKey, bot, event); err != nil {
			log.Printf("not call end DatetimePickerAction: %v", err)
		}
		return nil
	}
	sche.End[endKey] = &event.Postback.Params.Datetime
	title, ok := sche.Title[titleKey]
	if !ok {
		log.Print("title is nothing in sche")
	}
	start, ok := sche.Start[startKey]
	if !ok {
		log.Print("start is nothing in sche")
	}
	end, ok := sche.End[endKey]
	if !ok {
		log.Print("end is nothing in sche")
	}
	add := &calendar.Event{
		Summary: *title,
		Start: &calendar.EventDateTime{
			DateTime: *start + ":00+09:00",
			TimeZone: "Asia/Tokyo",
		},
		End: &calendar.EventDateTime{
			DateTime: *end + ":30+09:00",
			TimeZone: "Asia/Tokyo",
		},
	}
	log.Printf("after title: %v", title)
	log.Printf("start datetime: %v", start)
	log.Printf("end datetime: %v", end)

	b, err := ioutil.ReadFile("client_credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.JWTConfigFromJSON(b, calendar.CalendarScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := config.Client(oauth2.NoContext)
	cal, err := calendar.New(client)

	r, err := cal.Events.Insert(os.Getenv("CALENDAR_ID"), add).Do()
	if err != nil {
		log.Printf("calendar insert error: %v", err)
	} else {
		log.Printf("result: %v", r)
	}
	reply := fmt.Sprintf("schedule title: %v\nstart time: %v\nend time:%v\nGoogle Calendar has been updated!", *title, *start, *end)

	_, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(reply)).Do()
	if err != nil {
		log.Print(err)
		return err
	}
	return nil
}

func main() {
	// HTTP Handlerの初期化(LINEBot)
	bot, err := linebot.New(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
	)
	if err != nil {
		log.Fatal(err)
	}

	sche := &Schedule{
		make(map[string]*string),
		make(map[string]*string),
		make(map[string]*string),
	}

	// 実際にRequestを受け取った時に処理を行うHandle関数を定義し、handlerに登録
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {

		// events is defined
		events, err := bot.ParseRequest(r)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				w.WriteHeader(400)
			} else {
				w.WriteHeader(500)
			}
			return
		}

		for _, event := range events {
			switch event.Type {
			case linebot.EventTypeMessage:
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					sche.Title[titleKey] = &message.Text
					err = DatetimeAction(startKey, bot, event)
					if err != nil {
						log.Print(err)
					}
				default:
					_, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("やめてください")).Do()
					if err != nil {
						log.Fatal(err)
					}
				}

			case linebot.EventTypePostback:
				if err = DatetimePB(bot, event, sche); err != nil {
					log.Print(err)
				}
			}
		}
	})

	// port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	// HTTPサーバの起動
	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))

}
