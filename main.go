package main

import (
	"context"
	"encoding/json"
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

func DefaultMessage(bot *linebot.Client, event *linebot.Event) error {
	reply := linebot.NewTextMessage("今日も志を忘れず頑張ってください！！")
	if _, err := bot.ReplyMessage(event.ReplyToken, reply).Do(); err != nil {
		log.Print(err)
		return err
	}
	return nil
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

/*func tokenCacheFile() (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(user.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir, url.QueryEscape("// ちょっとわからない")), err
	// ちょっとわからない = "generate-schedule-calendar.json"
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatal(err)
	}
	return tok
}

func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.Create(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatal(err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

func NewGoogleAPI() (*calendar.Service, error) {
	ctx := context.Background()
	json, err := ioutil.ReadFile("client-secret.json")
	if err != nil {
		log.Fatal(err)
	}
	config, err := google.ConfigFromJSON(json, calendar.CalendarScope)
	if err != nil {
		log.Fatal(err)
	}
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatal(err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	client := config.Client(ctx, tok)
	NewCalen, err := calendar.New(client)
	if err != nil {
		log.Fatal(err)
	}
	return NewCalen, nil
}*/

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
		log.Fatalf("Unable to read client-secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

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
