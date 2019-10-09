package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/honeycombio/libhoney-go"
	"github.com/honeycombio/libhoney-go/transmission"
)

type DiscordWebhookPayload struct {
	Content string `json:"content"`
}

func PostDiscordWebhook() (bool, error) {
	discordURL, err := os.LookupEnv("DISCORD_WEBHOOK_URL")

	if err == false {
		panic("on no")
	}

	discordPayload := DiscordWebhookPayload{
		Content: "Testing from netlify",
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

func handler(request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
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
