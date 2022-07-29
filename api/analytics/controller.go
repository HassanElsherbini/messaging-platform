package analytics

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/HassanElsherbini/messaging-platform/services"
)

type Controller struct {
	messageService services.MessageService
}

func NewAnalyticsController(messageService services.MessageService) *Controller {
	return &Controller{
		messageService: messageService,
	}
}

func (c *Controller) Retrieve(w http.ResponseWriter, r *http.Request) {
	analytics, err := c.messageService.RetrieveAnalytics()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		log.Printf("Failed to retrieve analytics: %s", err)
	}

	result, err := json.Marshal(analytics)
	if err != nil {
		log.Printf("Unmarshal request body: %s", err)
		http.Error(w, "inernal service error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(result)
}
