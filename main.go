package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/logging"
	"github.com/line/line-bot-sdk-go/linebot"
)

type PostRequest struct {
	PostContent string `json:"post"`
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

		// gcloud logging
		ctx := context.Background()
		projectID := os.Getenv("PROJECT_ID")
		// logging client 初期化
		client, err := logging.NewClient(ctx, projectID)
		if err != nil {
			log.Fatalf("Failed to create client: %v", err)
		}
		defer client.Close()
		// Sets the name of the log to write to.
		logName := "calen-log"
		logger := client.Logger(logName).StandardLogger(logging.Info)
		// Stackdriver Logs
		logger.Println("hello world")
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
					case "あ":
						reply := linebot.NewTextMessage(message.Text)
						if _, err := bot.ReplyMessage(event.ReplyToken, reply).Do(); err != nil {
							log.Print(err)
						}
					case "カレン":
						log.Printf("%v\n", "Hello Calen !!")
						reply := linebot.NewTemplateMessage(
							"this is a botton template",
							linebot.NewButtonsTemplate(
								"https://shunsuarez.com/calendar.jpg",
								"Calendar",
								"Please select datetime",
								linebot.NewDatetimePickerAction("Make an appointment", "Datetime", "datetime", "", "", ""),
							),
						)
						if _, err = bot.ReplyMessage(event.ReplyToken, reply).Do(); err != nil {
							log.Print(err)
						}
					}
				}
			case linebot.EventTypePostback:
				data := event.Postback.Data
				log.Printf("here is postback %v\n", data)
				log.Printf("datetime is %v\n", data)
				_, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(data)).Do()
				if err != nil {
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
