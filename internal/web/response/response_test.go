package response

import (
	"encoding/json"
	"testing"
)

func TestOkResponse(t *testing.T) {
	t.Run("Set status code", func(t *testing.T) {
		b := OkResponse().
			SetStatusCode(200).
			Build()

		if b.StatusCode != 200 {
			t.Error("Status code not set")
		}

		if b.Body == nil {
			t.Error("Body not set")
		}
	})

	t.Run("Json body", func(t *testing.T) {
		type testBody struct {
			Test string `json:"test"`
		}

		body := testBody{
			Test: "test",
		}

		b := OkResponse().
			SetJSON(body).
			Build()

		if b.StatusCode != 200 {
			t.Error("Status code not set")
		}

		if b.Body == nil {
			t.Error("Body not set")
		}

		{
			var newB testBody
			err := json.Unmarshal(b.Body, &newB)
			if err != nil {
				t.Error("Failed to unmarshal body")
			}

			if newB.Test != "test" {
				t.Error("Body not set correctly")
			}
		}
	})

	t.Run("String body", func(t *testing.T) {
		b := OkResponse().
			SetBody("test").
			Build()

		if b.StatusCode != 200 {
			t.Error("Status code not set")
		}

		if b.Body == nil {
			t.Error("Body not set")
		}

		if string(b.Body) != "\"test\"" {
			t.Error("Body not set correctly")
		}
	})

	t.Run("bad formatted json", func(t *testing.T) {
		b := OkResponse().
			SetJSON(map[string]interface{}{
				"foo": make(chan int),
			}).
			Build()

		if b.StatusCode != 500 {
			t.Error("Status code not set")
		}

		if b.Body == nil {
			t.Error("Body not set")
		}

		if string(b.Body) != "Internal Server Error" {
			t.Error("Body not set correctly")
		}
	})
}
