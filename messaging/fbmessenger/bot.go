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
	"strconv"
	"time"

	"github.com/HassanElsherbini/messaging-platform/models"
	"github.com/HassanElsherbini/messaging-platform/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	sendMessageURI      string

	messageService services.MessageService
}

func NewBot(messageService services.MessageService, appSecret string, verifyToken string, accessToken string) *Bot {
	b := &Bot{
		messageService:      messageService,
		appSecret:           appSecret,
		verifyToken:         verifyToken,
		accessToken:         accessToken,
		sendMessageEndPoint: "https://graph.facebook.com/v14.0/me/messages",
	}

	b.sendMessageURI = fmt.Sprintf("%s?access_token=%s", b.sendMessageEndPoint, b.accessToken)

	return b
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

	go b.processMessageEvents(retrieveMessageEvents(event))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("EVENT_RECEIVED"))
}

func (b *Bot) processMessageEvents(messageEvents []MessageEvent) {
	for _, messageEvent := range messageEvents {
		if messageEvent.Read != nil {
			go b.processRead(messageEvent)
		} else if messageEvent.Feedback != nil {
			go b.processFeedback(messageEvent)
		}
	}
}

func (b *Bot) processRead(readEvent MessageEvent) {
	readAt := time.Unix(0, readEvent.Read.Watermark*int64(time.Millisecond))

	if err := b.messageService.MarkMessagesAsRead(readEvent.Sender.ID, readAt); err != nil {
		log.Printf("failed to mark messages as read: %s", err)
	}

}

func (b *Bot) processFeedback(feedbackEvent MessageEvent) {
	feedbacks := feedbackEvent.Feedback.FeedbackScreens[0].Questions

	feedbackIDs := []string{}
	for id := range feedbacks {
		feedbackIDs = append(feedbackIDs, id)
	}

	id := feedbackIDs[0]
	feedback := feedbacks[id]
	score, err := strconv.Atoi(feedback.Payload)
	if err != nil {
		log.Printf("failed to record response: %s", err)
		return
	}

	messageResponse := &models.MessageResponse{
		Score: models.MessageResponseScore{Value: score, Range: 5},
		Text:  feedback.FollowUp.Payload,
	}

	if err := b.messageService.AddMessageResponse(id, messageResponse); err != nil {
		log.Printf("failed to add message response: %s", err)
		return
	}
}

func retrieveMessageEvents(event Event) []MessageEvent {
	var events []MessageEvent
	for _, entry := range event.Entries {
		events = append(events, entry.Messaging...)
	}

	res, _ := json.MarshalIndent(events, "", " ")

	fmt.Println("RETRIEVE MSGS", string(res))
	return events
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
		log.Printf("Decode request body: %s", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if err := validateSendMessageRequest(incomingSendReq); err != nil {
		log.Printf("validate send message req: %s", err)
		http.Error(w, fmt.Sprintf("bad request: %s", err.Error()), http.StatusBadRequest)
		return
	}

	msgId := b.messageService.NewId()
	payload := NewCustomerFeedbackRequest(msgId.Hex(), incomingSendReq.RecipientID)

	_, err := b.sendMessage(payload)
	if err != nil {
		log.Printf("Send message: %s", err)
		http.Error(w, "failed to send message", http.StatusInternalServerError)
		return
	}

	newMessage := messageModelResolver(msgId, payload, incomingSendReq.TemplateType)
	fmt.Printf("%v", newMessage)

	if _, err := b.messageService.CreateMessage(newMessage); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		log.Printf("Failed to create new message: %s", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (b *Bot) sendMessage(payload *sendMessageRequest) ([]byte, error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	response, err := http.Post(b.sendMessageURI, "application/json", bytes.NewBuffer(payloadJSON))
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

func messageModelResolver(messageID primitive.ObjectID, fbMessage *sendMessageRequest, messageType string) *models.Message {
	return &models.Message{
		ID:           messageID,
		Channel:      "fbmessenger",
		RecipientID:  fbMessage.Recipient.ID,
		TemplateType: messageType,
		Body:         fbMessage.Message,
	}
}
