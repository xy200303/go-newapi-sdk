package main

import (
	"fmt"
	"log"
	"net/url"

	newapi "github.com/xy200303/go-newapi-sdk/newapi"
)

func main() {
	client, err := newapi.NewApiClient(
		"https://newapi.infinitext.cn",
		newapi.WithAdminAuth("20ZWNIKH+npUFYFEnDQ5LE4bjF1nOGMc", 1),
	)
	if err != nil {
		log.Fatal(err)
	}

	params := url.Values{}
	params.Set("p", "0")
	params.Set("page_size", "10")

	logs, err := client.AdminGetLogs(params)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("total=%d page=%d page_size=%d\n", logs.Total, logs.Page, logs.PageSize)
	for _, item := range logs.Items {
		fmt.Printf(
			"log id=%d user=%s model=%s quota=%d created_at=%d\n",
			item.ID,
			item.Username,
			item.ModelName,
			item.Quota,
			item.CreatedAt,
		)
	}
}
