package twitch

import "github.com/JoachimFlottorp/yeahapi/protobuf/collector"

type Chatters struct {
	Amount   int   `json:"total"`
	Chatters *collector.ChatterList `json:"chatters"`
}

func ChattersDefault() Chatters {
	return Chatters{
		Amount: 0,
		Chatters: &collector.ChatterList{
			Broadcaster: []string{},
			Vips:        []string{},
			Moderators:  []string{},
			Staff:       []string{},
			Viewers:     []string{},
		},
	}
}