package response

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// swagger:model apiResponse
type ApiResponse struct {
	Success bool `json:"success"`
	// Your request ID
	RequestID uuid.UUID `json:"request_id"`
	// The time the request was done
	Timestamp time.Time `json:"timestamp"`
	// The data of the request
	// This is dependent on the endpoint
	// refer to the endpoint documentation
	Data json.RawMessage `json:"data"`
	// Present if the request failed, otherwise null
	Error string `json:"error"`
}
