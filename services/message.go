package services

import (
	"context"
	"strings"
	"time"

	"github.com/HassanElsherbini/messaging-platform/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MessageService struct {
	messageCollection *mongo.Collection
	ctx               context.Context
}

type Analytics struct {
	Total     Stat             `json:"total"`
	ByDay     map[string]*Stat `json:"byDay"`
	Sentiment Sentiment        `json:"sentiment"`
}

type Stat struct {
	Sent             int `json:"sent"`
	Read             int `json:"read"`
	ReceivedResponse int `json:"replied"`
}

type Sentiment struct {
	Satisfied    int `json:"satisfied"`
	Neutral      int `json:"neutral"`
	UnSatisified int `json:"unSatisified"`
}

func NewMessageService(ctx context.Context, messagecCollection *mongo.Collection) MessageService {
	return MessageService{
		ctx:               ctx,
		messageCollection: messagecCollection,
	}
}

func (ms *MessageService) CreateMessage(message *models.Message) (string, error) {
	message.CreatedAt = time.Now()
	result, err := ms.messageCollection.InsertOne(context.TODO(), message)
	if err != nil {
		return "", err
	}

	id := result.InsertedID.(primitive.ObjectID).Hex()
	return id, nil
}

func (ms *MessageService) AddMessageResponse(messagID string, response *models.MessageResponse) error {
	id, err := primitive.ObjectIDFromHex(messagID)
	if err != nil {
		return err
	}

	response.CreatedAt = time.Now()
	update := bson.M{"$set": bson.M{
		"response": response,
	}}

	_, err = ms.messageCollection.UpdateByID(context.TODO(), id, update)

	if err != nil {
		return err
	}

	return nil
}

func (ms *MessageService) MarkMessagesAsRead(recipientID string, readAt time.Time) error {
	filter := bson.M{"recipient_id": recipientID, "read_at": time.Time{}}

	filter["created_at"] = bson.M{
		"$lte": readAt,
	}

	update := bson.M{"$set": bson.M{
		"read_at": readAt,
	}}

	_, err := ms.messageCollection.UpdateMany(context.TODO(), filter, update)

	if err != nil {
		return err
	}

	return nil
}

func (ms *MessageService) NewId() primitive.ObjectID {
	return primitive.NewObjectID()
}

// ANALYTICS
func (ms *MessageService) RetrieveAnalytics() (*Analytics, error) {
	opts := options.Find().SetProjection(bson.M{"read_at": 1, "created_at": 1, "response": 1})

	cursor, err := ms.messageCollection.Find(context.TODO(), bson.D{}, opts)

	if err != nil {
		return nil, err
	}

	var messages []models.Message
	if err = cursor.All(context.TODO(), &messages); err != nil {
		return nil, err
	}

	byDay := createByDayStats(messages)
	totals := createTotalStat(messages)
	sentiment := createSentimentStat(messages)

	return &Analytics{
		ByDay:     byDay,
		Total:     totals,
		Sentiment: sentiment,
	}, nil

}

func createByDayStats(messages []models.Message) map[string]*Stat {
	byDay := make(map[string]*Stat)

	for _, msg := range messages {
		sentDay := strings.ToLower(msg.CreatedAt.Weekday().String())

		if stat, ok := byDay[sentDay]; !ok {
			byDay[sentDay] = &Stat{Sent: 1}
		} else {
			stat.Sent++
		}

		if !msg.ReadAt.IsZero() {
			readDay := strings.ToLower(msg.ReadAt.Weekday().String())
			if stat, ok := byDay[readDay]; !ok {
				byDay[readDay] = &Stat{Read: 1}
			} else {
				stat.Read++
			}
		}

		if msg.Response != nil {
			respDay := strings.ToLower(msg.Response.CreatedAt.Weekday().String())
			if stat, ok := byDay[respDay]; !ok {
				byDay[respDay] = &Stat{ReceivedResponse: 1}
			} else {
				stat.ReceivedResponse++
			}
		}

	}

	return byDay
}

func createTotalStat(messages []models.Message) Stat {
	stat := Stat{}
	stat.Sent = len(messages)
	for _, msg := range messages {
		if !msg.ReadAt.IsZero() {
			stat.Read++
		}

		if msg.Response != nil {
			stat.ReceivedResponse++
		}
	}

	return stat
}

func createSentimentStat(messages []models.Message) Sentiment {
	sentiment := Sentiment{}
	for _, msg := range messages {
		resp := msg.Response
		if resp == nil || resp.Score == nil {
			continue
		}
		val := float64(resp.Score.Value) / float64(resp.Score.Range)

		switch {
		case val >= 0.8:
			sentiment.Satisfied++
		case val >= 0.6:
			sentiment.Neutral++
		default:
			sentiment.UnSatisified++
		}
	}

	return sentiment
}
