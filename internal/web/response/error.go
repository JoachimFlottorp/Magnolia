package response

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
)

type errorResponse struct {
	statusCode 	int
	err       	string
	headers   	map[string]string
	id			uuid.UUID
}

func Error() ErrorResponseBuilder {
	return &errorResponse{
		statusCode: 500,
		headers:    make(map[string]string),
		err:		http.StatusText(http.StatusInternalServerError),
		id: 		uuid.New(),
	}
}

func (b *errorResponse) SetStatusCode(statusCode int) RouterResponseBuilder {
	b.statusCode = statusCode
	return b
}

func (b *errorResponse) SetBody(body string) RouterResponseBuilder {
	b.err = body
	return b
}

func (b *errorResponse) SetJSON(body interface{}) RouterResponseBuilder {
	// Uhhh
	return b
}

func (b *errorResponse) SetCustomReqID(id uuid.UUID) RouterResponseBuilder {
	b.id = id
	return b
}

func (b *errorResponse) SetHeader(key, value string) RouterResponseBuilder {
	b.headers[key] = value
	return b
}

func (b *errorResponse) Build() RouterResponse {
	if b.err == "" {
		b.err = http.StatusText(b.statusCode)
	}

	return RouterResponse{
		StatusCode: b.statusCode,
		Headers:    b.headers,
		Error:      errors.New(b.err),
		UUID: 		b.id,
	}
}

func (b *errorResponse) InternalServerError(message ...string) ErrorResponseBuilder {
	b.statusCode = http.StatusInternalServerError
	b.err = getBodyOrCode(b.statusCode, message...)
	return b
}

func (b *errorResponse) NotFound(message ...string) ErrorResponseBuilder {
	b.statusCode = http.StatusNotFound
	b.err = getBodyOrCode(b.statusCode, message...)
	return b
}

func (b *errorResponse) BadRequest(message ...string) ErrorResponseBuilder {
	b.statusCode = http.StatusBadRequest
	b.err = getBodyOrCode(b.statusCode, message...)
	return b
}

func getBodyOrCode(code int, body ...string) string {
	if body != nil {
		return body[0]
	}
	return http.StatusText(code)
}