package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	ID           primitive.ObjectID     `bson:"_id,omitempty" json:"id,omitempty"`
	Channel      string                 `bson:"channel" json:"channel,omitempty" validate:"required"`
	RecipientID  string                 `bson:"recipient_id" json:"recipientID"`
	TemplateType string                 `bson:"template_type" json:"templateType,omitempty" validate:"required"`
	Body         map[string]interface{} `bson:"body" json:"body,omitempty" validate:"required"`
	Response     *MessageResponse       `bson:"response" json:"response"`
	ReadAt       time.Time              `bson:"read_at" json:"readAt"`
	CreatedAt    time.Time              `bson:"created_at" json:"createdAt,omitempty" validate:"required"`
}

type MessageResponse struct {
	Score     *MessageResponseScore `bson:"score" json:"score"`
	Text      string                `bson:"text" json:"text"`
	CreatedAt time.Time             `bson:"created_at" json:"createdAt,omitempty" validate:"required"`
}

type MessageResponseScore struct {
	Range int `bson:"range" json:"range,omitempty" validate:"required"`
	Value int `bson:"value" json:"value,omitempty" validate:"required"`
}
