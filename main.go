package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/line/line-bot-sdk-go/linebot"
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
			"Calendar",
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

func createEvent(datetime string, title string) *calendar.Event {
	event := &calendar.Event{
		Summary:  title,
		Location: "東京",
		Start: &calendar.EventDateTime{
			DateTime: datetime,
			TimeZone: "Asia/Tokyo",
		},
		/*End: &calendar.EventDateTime{
			Datetime: ,
			TimeZone: ,
		}*/
	}
	return event
}

func PostBack(bot *linebot.Client, event *linebot.Event) error {
	datetime := event.Postback.Params.Datetime
	log.Printf("here is postback %v\n", datetime)
	b, err := ioutil.ReadFile("client_credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarEventsScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	ctx := context.Background()
	ts := google.AppEngineTokenSource(ctx, calendar.CalendarEventsScope)
	t, err := ts.Token()
	if err != nil {
		log.Fatalf("Unable to create token from token source:%v", err)
	}
	client := config.Client(ctx, t)
	calendar, err := calendar.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	calendarEvent := createEvent(datetime, "未定")
	_, err = calendar.Events.Insert(os.Getenv("CALENDAR_ID"), calendarEvent).Do()
	if err != nil {
		log.Fatal(err)
	}

	// Throw a request here
	// POST https://www.googleapis.com/calendar/v3/calendars/calendarId/acl

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
