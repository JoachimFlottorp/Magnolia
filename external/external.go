package external

import (
	"net/http"
	"time"
)

var client = &http.Client{}

func init() {
	client = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     30 * time.Second,
		},
	}
}

func Client() *http.Client {
	return client
}
