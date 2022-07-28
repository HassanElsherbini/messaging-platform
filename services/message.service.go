package services

import (
	"context"

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

func (ms *MessageService) NewId() primitive.ObjectID {
	return primitive.NewObjectID()
}
