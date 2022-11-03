package client

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testClient(cb func(r *http.Request) (*http.Response, error)) *Client {
	return NewClient(Options{
		Transport: Middleware(cb),
	})
}

func TestClient(t *testing.T) {
	c := testClient(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       nil,
		}, nil
	})

	req, err := http.NewRequest("GET", "https://example.com", nil)

	assert.Nil(t, err)

	resp, err := c.Do(req)

	assert.Nil(t, err)

	body := make([]byte, 0)
	resp.Body.Read(body)

	assert.Equal(t, body, []byte{})
	assert.Equal(t, 200, resp.StatusCode)
}

func TestJSON(t *testing.T) {
	type testData struct {
		Foo string `json:"foo"`
		Bar string `json:"bar"`
	}

	testCases := []testData{
		{
			Foo: "foo",
			Bar: "bar",
		},
		{
			Foo: "baz",
			Bar: "qux",
		},
	}

	for _, tc := range testCases {
		c := testClient(func(r *http.Request) (*http.Response, error) {
			return JSONResponseTest(200, tc)
		})

		req, err := http.NewRequest("GET", "https://example.com", nil)

		assert.Nil(t, err)

		resp, err := c.Do(req)

		assert.Nil(t, err)

		body, err := DoJSON[testData](resp)

		assert.Nil(t, err)

		assert.Equal(t, tc, *body)
	}
}

func TestDoJSON(t *testing.T) {
	type testResp struct {
		Foo string `json:"foo"`
	}

	testCases := []struct {
		ShouldFail bool
		Body       []byte
	}{
		{
			ShouldFail: false,
			Body:       []byte(`{"foo": "bar"}`),
		},
		{
			ShouldFail: true,
			Body:       []byte(`{"foo": "bar"`),
		},
	}

	for _, tc := range testCases {
		resp := &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader(tc.Body)),
		}

		_, err := DoJSON[testResp](resp)

		if tc.ShouldFail {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestJSONResponseTest(t *testing.T) {
	type testData struct {
		Foo interface{} `json:"foo"`
	}

	testCases := []struct {
		ShouldFail bool
		Body       testData
	}{
		{
			ShouldFail: false,
			Body: testData{
				Foo: "bar",
			},
		},
		{
			ShouldFail: true,
			Body: testData{
				Foo: make(chan int),
			},
		},
	}

	for _, tc := range testCases {
		resp, err := JSONResponseTest(200, tc.Body)

		if tc.ShouldFail {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)

			body, err := DoJSON[testData](resp)

			assert.Nil(t, err)

			assert.Equal(t, tc.Body, *body)
		}
	}
}
