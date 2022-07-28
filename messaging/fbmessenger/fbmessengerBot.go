package fbmessenger

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/HassanElsherbini/messaging-platform/services"
)

var (
	// authorization errors
	errMissingHeaderSignature = errors.New("missing x-sign header singature")
	errInvalidHeaderSignature = errors.New("invalid x-sign header signature")

	// token verification errors
	errInvalidVerificationToken = errors.New("invalid verification token")
)

type Bot struct {
	appSecret           string
	verifyToken         string
	accessToken         string
	sendMessageEndPoint string

	messageService services.MessageService
}

func NewBot(messageService services.MessageService, appSecret string, verifyToken string, accessToken string) *Bot {
	return &Bot{
		messageService:      messageService,
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

func (b *Bot) Send(w http.ResponseWriter, r *http.Request) {
	var incomingSendReq incomingSendMessageRequest

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&incomingSendReq); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		log.Printf("Decode request body: %s", err)
		return
	}

	if err := validateSendMessageRequest(incomingSendReq); err != nil {
		http.Error(w, fmt.Sprintf("bad request: %s", err.Error()), http.StatusBadRequest)
		log.Printf("validate send message req: %s", err)
		return
	}

	msgId := b.messageService.NewId()
	payload := NewCustomerFeedbackRequest(msgId.Hex(), incomingSendReq.RecipientID)

	_, err := b.sendMessage(payload)
	if err != nil {
		http.Error(w, "failed to send message", http.StatusInternalServerError)
		log.Printf("Send message: %s", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (b *Bot) sendMessage(payload *sendMessageRequest) ([]byte, error) {
	url := fmt.Sprintf("%s?access_token=%s", b.sendMessageEndPoint, b.accessToken)

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	response, err := http.Post(url, "application/json", bytes.NewBuffer(payloadJSON))
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed with response code: %d. response: %s", response.StatusCode, string(responseBody))
	}

	return responseBody, nil
}

func validateSendMessageRequest(req incomingSendMessageRequest) error {
	if req.RecipientID == "" {
		return errors.New("missing recepient id")
	}

	if req.TemplateType != "customer_feedback" {
		return errors.New("invalid message template type")
	}

	return nil
}

func (b *Bot) sign(data []byte) []byte {
	hash := hmac.New(sha256.New, []byte(b.appSecret))
	hash.Reset()
	hash.Write(data)

	return hash.Sum(nil)
}

func NewCustomerFeedbackRequest(messageID string, recipientID string) *sendMessageRequest {
	message := map[string]interface{}{
		"attachment": map[string]interface{}{
			"type": "template",
			"payload": map[string]interface{}{
				"template_type": "customer_feedback",
				"title":         "Rate your recent shopping experience.",
				"subtitle":      "Let us know how we are doing by answering two questions",
				"button_title":  "Rate Experience",
				"feedback_screens": []interface{}{
					map[string]interface{}{
						"questions": []interface{}{
							map[string]interface{}{
								"id":           messageID,
								"type":         "csat",
								"title":        "How would you rate your recent shopping experience with us?",
								"score_label":  "neg_pos",
								"score_option": "five_stars",
								"follow_up": map[string]interface{}{
									"type":        "free_form",
									"placeholder": "Give additional feedback",
								},
							},
						},
					},
				},
				"business_privacy": map[string]interface{}{
					"url": "https://www.example.com",
				},
			},
		},
	}

	return &sendMessageRequest{
		MessagingType: "MESSAGE_TAG",
		Tag:           "CUSTOMER_FEEDBACK",
		Recipient:     MessageRecipient{ID: recipientID},
		Message:       message,
	}
}
