package main

import (
	"fmt"
	"log"
	"time"
)

type config struct {
	url string
}

func main() {
	client := NewClient(5 * time.Second)
	totalDeposits, err := client.GetTotalDeposits()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(totalDeposits)
}
