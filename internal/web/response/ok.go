package response

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

// Implements RouterResponseBuilder
type okResponseBuilder struct {
	statusCode 	int
	body 		json.RawMessage
	headers    	map[string]string
	isJson     	bool
}

func OkResponse() RouterResponseBuilder {
	return &okResponseBuilder{
		statusCode: 200,
		headers:    make(map[string]string),
		body: 	 	nil,
		isJson: 	false,
	}
}

func (b *okResponseBuilder) SetStatusCode(statusCode int) RouterResponseBuilder {
	b.statusCode = statusCode
	return b
}

func (b *okResponseBuilder) SetBody(body string) RouterResponseBuilder {
	b.body = json.RawMessage(fmt.Sprintf("\"%s\"", body))
	return b
}

func (b *okResponseBuilder) SetJSON(body interface{}) RouterResponseBuilder {
	marshalBody, err := json.Marshal(body)

	if err != nil {
		zap.S().Errorw("Failed to marshal JSON", "error", err)
		b.body = json.RawMessage("Internal Server Error")
		b.statusCode = 500
	} else {
		b.SetHeader("Content-Type", "application/json")
		b.body = marshalBody
		b.isJson = true
	}

	return b
}

func (b *okResponseBuilder) SetHeader(key, value string) RouterResponseBuilder {
	b.headers[key] = value
	return b
}

func (b *okResponseBuilder) Build() RouterResponse {
	if b.body == nil {
		b.body = json.RawMessage(http.StatusText(b.statusCode))
	}
	
	return RouterResponse {
		StatusCode: b.statusCode,
		Headers:    b.headers,
		Body: 	 	b.body,
	}
}