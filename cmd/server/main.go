package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	godotenv.Load(".env")
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is not set")
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("front")))

	mux.HandleFunc("/auth", cfg.HandlerAuth)
	mux.HandleFunc("/portfolio", cfg.handlerGetPortfolio)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving on: http://localhost:%s/\n", port)
	log.Fatal(srv.ListenAndServe())
}
