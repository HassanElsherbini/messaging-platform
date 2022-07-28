package fbmessenger

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

var (
	// authorization errors
	errMissingHeaderSignature = errors.New("missing x-sign header singature")
	errInvalidHeaderSignature = errors.New("invalid x-sign header signature")

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

func (b *Bot) Receive(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		log.Printf("Reading request payload: %s", err)
		return
	}

	if err := b.validateSignature(payload, r.Header.Get("X-Hub-Signature-256")); err != nil {
		log.Printf("Validate signature: %s", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	event := Event{}
	if err := json.Unmarshal(payload, &event); err != nil {
		log.Printf("Unmarshal request body: %s", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("EVENT_RECEIVED"))
}

func (b *Bot) validateSignature(data []byte, signature string) error {
	if len(signature) == 0 {
		return errMissingHeaderSignature
	}

	actualSignature, err := hex.DecodeString(signature[7:])
	if err != nil {
		return fmt.Errorf("decode signature: %w", err)
	}

	expectedSignature := b.sign(data)
	if !hmac.Equal(expectedSignature, actualSignature) {
		return errInvalidHeaderSignature
	}

	return nil
}

func (b *Bot) sign(data []byte) []byte {
	hash := hmac.New(sha256.New, []byte(b.appSecret))
	hash.Reset()
	hash.Write(data)

	return hash.Sum(nil)
}
