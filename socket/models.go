package socket

import "time"

// EventType defines the type of data change event
type EventType string

const (
	EventCreated EventType = "created"
	EventUpdated EventType = "updated"
	EventDeleted EventType = "deleted"
)

// DataChangeEvent represents a change in our data that clients should know about
type DataChangeEvent struct {
	Type      EventType `json:"type"`
	Entity    string    `json:"entity"`
	Data      any       `json:"data"`
	Timestamp time.Time `json:"timestamp"`
}
