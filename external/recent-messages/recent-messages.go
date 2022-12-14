/*
	This package is responsible for querying https://github.com/robotty/recent-messages2 for recent messages.

	This allows me to display recent messages from a channel in https://chatterino.com/.
*/

package recentmessages

import (
	"sync"

	"go.uber.org/zap"
)

type EndpointName string

const (
	/* The one i personally use */
	EndpointSnakes = EndpointName("sneiks")
)

type Endpoint interface {
	Url() string
	MakeRequest(string) error
}

var endpoints = map[EndpointName]Endpoint{
	EndpointSnakes: snakes(),
}

func Request(endpoint EndpointName, channels []string) {
	wg := &sync.WaitGroup{}

	if e, ok := endpoints[endpoint]; ok {
		for _, channel := range channels {
			wg.Add(1)
			go func(channel string) {
				defer wg.Done()

				err := e.MakeRequest(channel)
				if err != nil {
					zap.S().Errorw("Failed to request recent messages", "error", err)
				}
			}(channel)
		}
	}

	wg.Wait()
}
