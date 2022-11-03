/*
	This package is responsible for querying https://github.com/robotty/recent-messages2 for recent messages.

	This allows me to display recent messages from a channel in https://chatterino.com/.
*/

package recentmessages

type EndpointName string

const (
	/* The one i personally use */
	EndpointSnakes = EndpointName("sneiks")
)

type Endpoint interface {
	Url() string
	MakeRequest([]string) error
}

var endpoints = map[EndpointName]Endpoint{
	EndpointSnakes: snakes(),
}

func Request(endpoint EndpointName, channels []string) error {
	if e, ok := endpoints[endpoint]; ok {
		return e.MakeRequest(channels)
	}
	return nil
}
