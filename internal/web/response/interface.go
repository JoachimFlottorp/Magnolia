package response

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// swagger:model apiResponse
type ApiResponse struct {
	// If the request was successful - always true
	Success 	bool       	`json:"success"`
	// Your request ID
	RequestID 	uuid.UUID 	`json:"request_id"`
	// The time the request was done
	Timestamp 	time.Time 	`json:"timestamp"`
	// The data of the request
	// This is dependent on the endpoint
	// refer to the endpoint documentation
	Data    json.RawMessage `json:"data,omitempty"`
	// Present if the request failed, otherwise null
	Error	string   `json:"error,omitempty"`
}

// RouterResponse defines the response of every handler
type RouterResponse struct {
	StatusCode int
	Headers    map[string]string
	Body 	   json.RawMessage
	Error 	   error
}

// RouterErrorResponse defines a builder for creating a RouterResponse
type RouterResponseBuilder interface {
	SetStatusCode(int) RouterResponseBuilder
	SetBody(string) RouterResponseBuilder
	SetJSON(interface{}) RouterResponseBuilder
	SetHeader(string, string) RouterResponseBuilder
	Build() RouterResponse
}

type ErrorResponseBuilder interface {
	RouterResponseBuilder

	InternalServerError(...string) ErrorResponseBuilder
	NotFound(...string) ErrorResponseBuilder
	BadRequest(...string) ErrorResponseBuilder
}