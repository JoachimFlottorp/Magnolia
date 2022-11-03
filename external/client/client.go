package client

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type Client struct {
	*http.Client
}

type Options struct {
	Transport http.RoundTripper
}

func NewClient(opts Options) *Client {
	t := http.DefaultTransport
	if opts.Transport != nil {
		t = opts.Transport
	}

	return &Client{
		Client: &http.Client{
			Transport: t,
		},
	}
}

func DoJSON[T interface{}](resp *http.Response) (*T, error) {
	defer func() {
		if resp.Body != nil {
			resp.Body.Close()
		}
	}()

	var b T
	if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
		return nil, err
	}

	return &b, nil
}

////// Test Related //////

type RoundTripper struct {
	interceptor func(r *http.Request) (*http.Response, error)
}

func (r RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return r.interceptor(req)
}

func Middleware(cb func(r *http.Request) (*http.Response, error)) http.RoundTripper {
	return &RoundTripper{
		interceptor: cb,
	}
}

func JSONResponseTest[T interface{}](code int, body T) (*http.Response, error) {
	req, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	header := http.Header{}
	header.Set("Content-Type", "application/json")

	return &http.Response{
		StatusCode: code,
		Body:       io.NopCloser(bytes.NewReader(req)),
		Header:     header,
	}, nil
}
