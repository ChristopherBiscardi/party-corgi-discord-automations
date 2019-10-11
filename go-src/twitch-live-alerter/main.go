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

// ZapierTwitchNotificationPayload is the twitch content payload
// from zapier
type ZapierTwitchNotificationPayload struct {
	UserID       string       `json:"user_id"`
	Language     string       `json:"language"`
	Title        string       `json:"title"`
	EventType    string       `json:"type"`
	StreamerInfo StreamerInfo `json:"streamer_info"`
	ThumbnailURL string       `json:"thumbnail_url"`
	GameID       string       `json:"game_id"`
	StartedAt    string       `json:"started_at"`
	UserName     string       `json:"user_name"`
	ID           string       `json:"id"`
	ViewerCount  string       `json:"viewer_count"`
}

// StreamerInfo is the Twitch streamer_info field
type StreamerInfo struct {
	ViewCount       string `json:"view_count"`
	OfflineImageURL string `json:"offline_image_url"`
	Description     string `json:"description"`
	ProfileImageURL string `json:"profile_image_url"`
	StreamURL       string `json:"stream_url"`
	Login           string `json:"login"`
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

func ReceiveTwitchPayload(request *events.APIGatewayProxyRequest) (ZapierTwitchNotificationPayload, error) {
	twitchPayload := ZapierTwitchNotificationPayload{}
	byteBody := []byte(request.Body)

	if err := json.Unmarshal(byteBody, &twitchPayload); err != nil {
		return twitchPayload, errors.New("JSON unmarshall failed")
	}
	return twitchPayload, nil
}
func handler(request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {

	zapierPayload, err := ReceiveTwitchPayload(&request)
	if err != nil {
		return &events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       "Hello, World",
		}, nil
	}
	PostDiscordWebhook(DiscordWebhookPayload{
		Content: fmt.Sprintf(
			"%s started streaming \"%s\" at %s",
			zapierPayload.UserName,
			zapierPayload.Title,
			zapierPayload.StreamerInfo.StreamURL,
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
