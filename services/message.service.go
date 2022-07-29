package services

import (
	"context"
	"time"

	"github.com/HassanElsherbini/messaging-platform/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MessageService struct {
	messageCollection *mongo.Collection
	ctx               context.Context
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

func (ms *MessageService) NewId() primitive.ObjectID {
	return primitive.NewObjectID()
}
