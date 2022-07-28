package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/HassanElsherbini/messaging-platform/messaging/fbmessenger"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		panic(err)
	}

	config, err := godotenv.Read(".env")
	if err != nil {
		panic(err)
	}

	fbBot := fbmessenger.NewBot(config["APP_SECRET"], config["VERIFY_TOKEN"], config["ACCESS_TOKEN"])

	router := mux.NewRouter()
	router.HandleFunc("/api/messaging/receive/fbmessenger", fbBot.Verify).Methods("GET")

	log.Printf("Server listening on port %v", config["SERVER_PORT"])
	if err := http.ListenAndServe(fmt.Sprintf(":%s", config["SERVER_PORT"]), router); err != nil {
		log.Fatal(err)
	}

}
