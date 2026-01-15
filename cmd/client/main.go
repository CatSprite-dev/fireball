package main

import (
	"log"
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	totalReturn, err := cfg.GetTotalReturn()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(totalReturn)
}
