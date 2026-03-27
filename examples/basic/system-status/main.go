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
	rootToken := os.Getenv("NEWAPI_ROOT_TOKEN")

	client, err := newapi.NewApiClient(
		baseURL,
		newapi.WithAdminAuth(rootToken, 1),
	)
	if err != nil {
		log.Fatal(err)
	}

	var status map[string]any
	err = client.Management.System.StatusGet.Do(context.Background(), nil, &status)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("system status: %+v\n", status)
}
