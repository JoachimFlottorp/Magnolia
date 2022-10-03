package mongo

import (
	"time"
)

type ApiLog struct {
	// ID is the unique identifier for the log entry.
	ID string `json:"id" bson:"_id"`

	// Timestamp is the time the log entry was created.
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`

	// Method is the HTTP method used for the request.
	Method string `json:"method" bson:"method"`

	// Path is the path used for the request.
	Path string `json:"path" bson:"path"`

	// Status is the HTTP status code returned for the request.
	Status int `json:"status" bson:"status"`

	// IP is the IP address of the client.
	IP string `json:"ip" bson:"ip"`

	// UserAgent is the user agent of the client.
	UserAgent string `json:"user_agent" bson:"user_agent"`

	// Error is the error returned for the request.
	Error string `json:"error,omitempty" bson:"error"`
}
