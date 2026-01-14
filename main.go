package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/CatSprite-dev/fireball/internal/api"
)

type congig struct {
	url string
}

func main() {
	client := api.NewClient(5 * time.Second)
	cfg := congig{
		url: *client.GetBaseURL(),
	}
	user, err := client.GetUserInfo(cfg.url)
	if err != nil {
		log.Printf("Error fetching user info: %s", err)
	}

	jsonData, err := json.MarshalIndent(user, "", "  ") // "" - префикс, "  " - отступ
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(jsonData))
}
