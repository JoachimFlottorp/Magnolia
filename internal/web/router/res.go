package router

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ApiOkay struct {
	Success 	bool       	`json:"success"`
	RequestID 	uuid.UUID 	`json:"request_id"`
	Timestamp 	time.Time 	`json:"timestamp"`
	
	Data    interface{} `json:"data"`
}

type ApiFail struct {
	Success 	bool       	`json:"success"`
	RequestID 	uuid.UUID 	`json:"request_id"`
	Timestamp 	time.Time 	`json:"timestamp"`

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
