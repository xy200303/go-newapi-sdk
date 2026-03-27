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

	var result map[string]any
	err = client.AIModel.Chat.OpenAI.CreateChatCompletionPost.Do(
		context.Background(),
		&newapi.CallConfig{
			JSONBody: map[string]any{
				"model": "gpt-4o-mini",
				"messages": []map[string]any{
					{"role": "system", "content": "You are a helpful assistant."},
					{"role": "user", "content": "你好，帮我写一句欢迎语"},
				},
			},
		},
		&result,
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("chat completion response: %+v\n", result)
}
