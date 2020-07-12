package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/line/line-bot-sdk-go/linebot"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

type PostRequest struct {
	PostContent string `json:"post"`
}

func DefaultMessage(bot *linebot.Client, event *linebot.Event) error {
	reply := linebot.NewTextMessage("今日も志を忘れず頑張ってください！！")
	if _, err := bot.ReplyMessage(event.ReplyToken, reply).Do(); err != nil {
		log.Print(err)
		return err
	}
	return nil
}

func CallCalen(bot *linebot.Client, event *linebot.Event) error {
	reply := linebot.NewTemplateMessage(
		"this is a botton template",
		linebot.NewButtonsTemplate(
			"https://shunsuarez.com/calendar.jpg",
			"Hi! I'm Calen",
			"Please select datetime",
			linebot.NewDatetimePickerAction("Make an appointment", "Datetime", "datetime", "", "", ""),
		),
	)
	_, err := bot.ReplyMessage(event.ReplyToken, reply).Do()
	if err != nil {
		log.Print(err)
		return err
	}
	return nil
}

func PostBack(bot *linebot.Client, event *linebot.Event) error {
	datetime := event.Postback.Params.Datetime

	b, err := ioutil.ReadFile("client_credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.JWTConfigFromJSON(b, calendar.CalendarScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := config.Client(oauth2.NoContext)
	cl, err := calendar.New(client)

	/*events, err := calendar.Events.List(os.Getenv("CALENDAR_ID")).Do()
	if len(events.Items) == 0 {
		log.Printf("No upcoming events found.")
	} else {
		for _, item := range events.Items {
			date := item.Start.DateTime
			if date == "" {
				date = item.Start.Date
			}
			fmt.Printf("%v (%v)\n", item.Summary, date)
		}
	}*/

	rand.Seed(time.Now().UnixNano())
	add := &calendar.Event{
		Summary: "未定",
		Start: &calendar.EventDateTime{
			DateTime: datetime + ":00+09:00",
			TimeZone: "Asia/Tokyo",
		},
		End: &calendar.EventDateTime{
			DateTime: datetime + ":30+09:00",
			TimeZone: "Asia/Tokyo",
		},
	}
	log.Printf("add event: %v", add)
	_, err = cl.Events.Insert(os.Getenv("CALENDAR_ID"), add).Do()
	if err != nil {
		log.Printf("calendar insert error: %v", err)
	}

	_, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(datetime)).Do()
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
					switch message.Text {
					case "カレン":
						if err = CallCalen(bot, event); err != nil {
							log.Print(err)
						}
					default:
						if err = DefaultMessage(bot, event); err != nil {
							log.Print(err)
						}
					}
				}
			case linebot.EventTypePostback:
				if err = PostBack(bot, event); err != nil {
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
