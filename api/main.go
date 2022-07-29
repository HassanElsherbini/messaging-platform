package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/HassanElsherbini/messaging-platform/analytics"
	"github.com/HassanElsherbini/messaging-platform/messaging/fbmessenger"
	"github.com/HassanElsherbini/messaging-platform/services"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		panic(err)
	}

	config, err := godotenv.Read(".env")
	if err != nil {
		panic(err)
	}

	clientOptions := options.Client().
		ApplyURI(config["MONGODB_URI"])

	client, err := mongo.Connect(context.TODO(), clientOptions)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(ctx)

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to mongo!")
	messageCollection := client.Database("messaging-platform").Collection("messages")

	messageService := services.NewMessageService(ctx, messageCollection)

	fbBot := fbmessenger.NewBot(messageService, config["APP_SECRET"], config["VERIFY_TOKEN"], config["ACCESS_TOKEN"])
	analyticsController := analytics.NewAnalyticsController(messageService)

	router := mux.NewRouter()
	router.HandleFunc("/api/messaging/receive/fbmessenger", fbBot.Verify).Methods("GET")
	router.HandleFunc("/api/messaging/receive/fbmessenger", fbBot.Receive).Methods("POST")
	router.HandleFunc("/api/messaging/send/fbmessenger", fbBot.Send).Methods("POST")

	router.HandleFunc("/api/analytics/", analyticsController.Retrieve).Methods("GET")

	log.Printf("Server listening on port %v", config["SERVER_PORT"])
	if err := http.ListenAndServe(fmt.Sprintf(":%s", config["SERVER_PORT"]), router); err != nil {
		log.Fatal(err)
	}

}
