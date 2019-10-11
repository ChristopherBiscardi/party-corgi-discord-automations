package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/honeycombio/libhoney-go"
	"github.com/honeycombio/libhoney-go/transmission"
)

// DiscordWebhookPayload is the content we send to
// discord
type DiscordWebhookPayload struct {
	Content string `json:"content"`
}

// TwitchNotificationPayload represents a webhook
// payload from twitch, which are nested in a
// `data` field
type TwitchNotificationPayload struct {
	Data []TwitchStreamChangeEvent `json:"data"`
}

// TwitchStreamChangeEvent is the payload in the array here:
// https://dev.twitch.tv/docs/api/webhooks-reference/#topic-stream-changed
type TwitchStreamChangeEvent struct {
	UserName string `json:"user_name"`
	Title    string `json:"title"`
}

// PostDiscordWebhook sends a webhook to discord
// with the supplied content
func PostDiscordWebhook(discordPayload DiscordWebhookPayload) (bool, error) {
	discordURL, err := os.LookupEnv("DISCORD_WEBHOOK_URL")

	if err == false {
		panic("on no")
	}

	s, _ := json.Marshal(discordPayload)
	b := bytes.NewBuffer(s)

	var myClient = &http.Client{Timeout: 10 * time.Second}
	r, postErr := myClient.Post(discordURL, "application/json", b)
	if postErr != nil {
		return false, postErr
	}
	defer r.Body.Close()

	return true, nil
}

func ReceiveTwitchPayload(request *events.APIGatewayProxyRequest) (TwitchStreamChangeEvent, error) {
	twitchPayload := TwitchNotificationPayload{}
	byteBody := []byte(request.Body)

	if err := json.Unmarshal(byteBody, &twitchPayload); err != nil {
		return TwitchStreamChangeEvent{}, errors.New("JSON unmarshall failed")
	}
	if len(twitchPayload.Data) > 0 {
		return twitchPayload.Data[0], nil
	}
	return TwitchStreamChangeEvent{}, errors.New("failed to decode a single twitch stream change event")
}
func handler(request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {

	twitchStreamChangeEvent, err := ReceiveTwitchPayload(&request)
	if err != nil {
		return &events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       "Hello, World",
		}, nil
	}
	PostDiscordWebhook(DiscordWebhookPayload{
		Content: fmt.Sprintf(
			"%s started streaming \"%s\" at https://twitch.tv/%s",
			twitchStreamChangeEvent.UserName,
			twitchStreamChangeEvent.Title,
			twitchStreamChangeEvent.UserName,
		),
	})

	return &events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "Hello, World",
	}, nil
}

func main() {
	// writeKey, _ := os.LookupEnv("HONEYCOMB_WRITE_KEY")
	libhoney.Init(libhoney.Config{
		// WriteKey: writeKey,
		Dataset:      "netlify-serverless",
		Transmission: &transmission.WriterSender{},
	})
	// Flush any pending calls to Honeycomb before exiting
	defer libhoney.Close()
	// Make the handler available for Remote Procedure Call by AWS Lambda
	lambda.Start(handler)
}
