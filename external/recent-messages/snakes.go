package recentmessages

import (
	"net/http"

	"github.com/JoachimFlottorp/magnolia/external"
)

const snakesUrl = "https://recent-messages.zneix.eu/"

var snakesClient = external.NewKeepAliveClient()

type s struct{}

func snakes() Endpoint {
	return &s{}
}

func (s *s) Url() string {
	return snakesUrl
}

func (s *s) MakeRequest(channels []string) error {
	req, err := http.NewRequest("GET", snakesUrl, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "Magnolia")

	_, err = snakesClient.Do(req)
	if err != nil {
		return err
	}

	return nil
}