package main

import (
	"context"
	"fmt"
	"log"
	"os"

	newapi "xy200303/go-newapi-sdk/newapi"
)

func main() {
	baseURL := os.Getenv("NEWAPI_BASE_URL")
	token := os.Getenv("NEWAPI_TOKEN")

	client, err := newapi.NewApiClient(
		baseURL,
		newapi.WithBearerToken(token),
	)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.AIModel.Audio.OpenAI.CreateSpeechPost.DoRaw(
		context.Background(),
		&newapi.CallConfig{
			JSONBody: map[string]any{
				"model": "tts-1",
				"input": "hello world",
				"voice": "alloy",
			},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Printf("status=%d content-type=%s\n", resp.StatusCode, resp.Header.Get("Content-Type"))
}
