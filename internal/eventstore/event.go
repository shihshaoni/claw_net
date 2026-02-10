package eventstore

import "time"

type EventType string

const (
	EventMessageSent     EventType = "MESSAGE_SENT"
	EventMessageReceived EventType = "MESSAGE_RECEIVED"
	EventStateTransition EventType = "STATE_TRANSITION"
	EventErrorRaised     EventType = "ERROR_RAISED"
	EventBudgetUpdated   EventType = "BUDGET_UPDATED"
)

type Event struct {
	EventID   string
	RunID     string
	SeqNo     int64
	Type      EventType
	Actor     string
	Data      any
	Timestamp time.Time
}
