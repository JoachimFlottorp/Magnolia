package response

import (
	"errors"
	"net/http"
)

type errorResponse struct {
	statusCode 	int
	err       	string
	headers   	map[string]string
}

func Error() ErrorResponseBuilder {
	return &errorResponse{
		statusCode: 500,
		headers:    make(map[string]string),
		err:		http.StatusText(http.StatusInternalServerError),
	}
}

func (b *errorResponse) SetStatusCode(statusCode int) RouterResponseBuilder {
	b.statusCode = statusCode
	return b
}

func (b *errorResponse) SetBody(body string) RouterResponseBuilder {
	// b.err = json.RawMessage(fmt.Sprintf("\"%s\"", body))
	b.err = body
	return b
}

func (b *errorResponse) SetJSON(body interface{}) RouterResponseBuilder {
	// Uhhh
	return b
}

func (b *errorResponse) SetHeader(key, value string) RouterResponseBuilder {
	b.headers[key] = value
	return b
}

func (b *errorResponse) Build() RouterResponse {
	if b.err == "" {
		b.err = http.StatusText(b.statusCode)
		// b.err = json.RawMessage(http.StatusText(b.statusCode))
	}

	return RouterResponse{
		StatusCode: b.statusCode,
		Headers:    b.headers,
		Error:      errors.New(b.err),
	}
}

func (b *errorResponse) InternalServerError(message ...string) ErrorResponseBuilder {
	b.statusCode = http.StatusInternalServerError
	b.err = getBodyOrCode(message...)
	// b.error = json.RawMessage(fmt.Sprintf("\"%s\"", getBodyOrCode(message...)))
	return b
}

func (b *errorResponse) NotFound(message ...string) ErrorResponseBuilder {
	b.statusCode = http.StatusNotFound
	b.err = getBodyOrCode(message...)
	// b.error = json.RawMessage(fmt.Sprintf("\"%s\"", getBodyOrCode(message...)))
	return b
}

func (b *errorResponse) BadRequest(message ...string) ErrorResponseBuilder {
	b.statusCode = http.StatusBadRequest
	b.err = getBodyOrCode(message...)
	// b.err = json.RawMessage(fmt.Sprintf("\"%s\"", getBodyOrCode(message...)))
	return b
}

func getBodyOrCode(body ...string) string {
	if body != nil {
		return body[0]
	}
	return http.StatusText(http.StatusBadRequest)
}