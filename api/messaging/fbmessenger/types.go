package fbmessenger

type Event struct {
	Object  string       `json:"object"`
	Entries []EventEntry `json:"entry"`
}

type EventEntry struct {
	ID        string         `json:"id"`
	Time      int            `json:"time"`
	Messaging []MessageEvent `json:"messaging"`
}

type MessageEvent struct {
	Sender    MessageSender    `json:"sender"`
	Recipient MessageRecipient `json:"recipient"`
	TimeStamp int              `json:"timestamp"`
	Message   *ReceivedMessage `json:"message"`

	Feedback *Feedback `json:"messaging_feedback"`

	Read *Read `json:"read"`
}

type MessageRecipient struct {
	ID string `json:"id"`
}

type MessageSender struct {
	ID string `json:"id"`
}

type ReceivedMessage struct {
	Mid  string `json:"mid,omitempty"`
	Seq  int    `json:"seq,omitempty"`
	Text string `json:"text"`
}

type Feedback struct {
	FeedbackScreens []FeedbackScreen `json:"feedback_screens"`
}

type FeedbackScreen struct {
	ScreenID  int                         `json:"screen_id"`
	Questions map[string]FeedbackQuestion `json:"questions"`
}

type FeedbackQuestion struct {
	Payload  string                    `json:"payload"`
	Type     string                    `json:"type"`
	FollowUp *FeedbackQuestionFollowUp `json:"follow_up"`
}

type FeedbackQuestionFollowUp struct {
	Payload string `json:"payload"`
	Type    string `json:"type"`
}

type Read struct {
	Watermark int64 `json:"watermark"`
}

type sendMessageRequest struct {
	MessagingType string                 `json:"messaging_type"`
	Recipient     MessageRecipient       `json:"recipient"`
	Tag           string                 `json:"tag,omitempty"`
	Message       map[string]interface{} `json:"message"`
}

type incomingSendMessageRequest struct {
	RecipientID  string `json:"recipientId"`
	TemplateType string `json:"templateType"`
	Payload      string `json:"payload"`
}
