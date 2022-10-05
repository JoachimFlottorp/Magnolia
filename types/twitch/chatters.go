package twitch

import "github.com/JoachimFlottorp/yeahapi/protobuf/collector"

// swagger:model
type Chatters struct {
	Total   int   `json:"total"`
	Chatters *collector.ChatterList `json:"chatters"`
}

func ChattersDefault() Chatters {
	return Chatters{
		Total: 0,
		Chatters: &collector.ChatterList{
			Broadcaster: []string{},
			Vips:        []string{},
			Moderators:  []string{},
			Staff:       []string{},
			Viewers:     []string{},
		},
	}
}