package router

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// swagger:model apiOkay
type ApiOkay struct {
	// If the request was successful - always true
	Success 	bool       	`json:"success"`
	// Your request ID
	RequestID 	uuid.UUID 	`json:"request_id"`
	// The time the request was done
	Timestamp 	time.Time 	`json:"timestamp"`
	
	// The data of the request
	// This is dependent on the endpoint
	// refer to the endpoint documentation
	Data    interface{} `json:"data"`
}

// swagger:model apiFail
type ApiFail struct {
	// If the request was successful - always false
	Success 	bool       	`json:"success"`
	// Your request ID
	RequestID 	uuid.UUID 	`json:"request_id"`
	// The time the request was done
	Timestamp 	time.Time 	`json:"timestamp"`

	// The error message
	Error  string `json:"error"`
}

func Send[T interface{}](w http.ResponseWriter, status int, msg T) {
	w.Header().Set("Content-Type", "application/json")
	j, err := json.MarshalIndent(msg, "", "  ")

	if err != nil {
		zap.S().Warnf("Failed to marshal JSON: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(status)
	w.Write(j)
}
