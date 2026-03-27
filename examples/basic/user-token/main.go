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
	username := os.Getenv("NEWAPI_USERNAME")
	password := os.Getenv("NEWAPI_PASSWORD")
	userID := 1

	client, err := newapi.NewApiClient(baseURL)
	if err != nil {
		log.Fatal(err)
	}

	sessionCookie, err := client.UserLoginContext(context.Background(), username, password)
	if err != nil {
		log.Fatal(err)
	}

	accessToken, err := client.UserGenerateAccessTokenContext(context.Background(), sessionCookie, userID)
	if err != nil {
		log.Fatal(err)
	}

	token, err := client.UserCreateTokenContext(
		context.Background(),
		accessToken,
		userID,
		newapi.CreateTokenRequest{
			Name:           "demo-token",
			RemainQuota:    0,
			UnlimitedQuota: true,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("created token: %+v\n", token)
}
