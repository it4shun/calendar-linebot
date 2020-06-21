package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/line/line-bot-sdk-go/linebot/httphandler"
)

func main() {
	// HTTP Handlerの初期化
	handler, err := httphandler.New(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
	)
	if err != nil {
		log.Fatal(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	// 実際にRequestを受け取った時に処理を行うHandle関数を定義し、handlerに登録
	handler.HandleEvents(func(events []*linebot.Event, r *http.Request) {
		bot, err := handler.NewClient()
		if err != nil {
			log.Print(err)
			return
		}

		for _, event := range events {
			if event.Type != linebot.EventTypeMessage {
				return
			}

			//switch message := event.Message.(type) {
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
						reply := linebot.NewTemplateMessage(
							"this is a botton template",
							linebot.NewButtonsTemplate(
								"https://shunsuarez.com/calendar.jpg",
								"Calendar",
								"Please select datetime",
								linebot.NewDatetimePickerAction("Make an appointment", "id=1", "datetime", "", "", ""),
							),
						)
						if _, err = bot.ReplyMessage(event.ReplyToken, reply).Do(); err != nil {
							log.Print(err)
						}
					}
				}
			case linebot.EventTypePostback:
				reply := linebot.NewTextMessage(string(event.Postback.Params.Datetime))
				if _, err = bot.ReplyMessage(event.ReplyToken, reply).Do(); err != nil {
					log.Print(err)
				}
			}
		}
	})

	// /callback にエンドポイントの定義
	http.Handle("/callback", handler)
	// HTTPサーバの起動
	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
