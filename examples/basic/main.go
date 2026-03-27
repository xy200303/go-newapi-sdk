package main

import (
	"context"
	"fmt"
	"log"

	newapi "xy200303/go-newapi-sdk/newapi"
)

func main() {
	client, err := newapi.NewApiClient(
		"https://api.example.com",
		newapi.WithAdminAuth("root-token", 1),
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
