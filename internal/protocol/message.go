package protocol

import "time"

type Message struct {
	MessageID string
	RunID     string
	TaskID    string

	FromAgent string
	ToAgents  []string

	Intent     IntentType
	Payload    any
	Confidence float64
	Refs       []string

	IdempotencyKey string
	Timestamp      time.Time
}
