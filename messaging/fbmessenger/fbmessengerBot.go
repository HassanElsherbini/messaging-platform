package fbmessenger

import (
	"errors"
	"log"
	"net/http"
)

var (
	// token verification errors
	errInvalidVerificationToken = errors.New("invalid verification token")
)

type Bot struct {
	appSecret   string
	verifyToken string
	accessToken string

	sendMessageEndPoint string
}

func NewBot(appSecret string, verifyToken string, accessToken string) *Bot {
	return &Bot{
		appSecret:           appSecret,
		verifyToken:         verifyToken,
		accessToken:         accessToken,
		sendMessageEndPoint: "https://graph.facebook.com/v14.0/me/messages",
	}
}

func (b *Bot) Verify(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("hub.verify_token")
	mode := r.URL.Query().Get("hub.mode")
	challenge := r.URL.Query().Get("hub.challenge")

	if mode == "subscribe" && token == b.verifyToken {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(challenge))
		log.Println("Webhook verified")
	} else {
		http.Error(w, "invalid token", http.StatusForbidden)
		log.Printf("Verifiy token: %s", errInvalidVerificationToken)
	}
}
