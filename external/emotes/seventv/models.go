package seventv

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GlobalEmoteResponse is the response from the 7TV API for global emotes.
// It is just an emoteset
type GlobalEmoteResponse struct {
	EmoteSet
}

// ChannelEmoteResponse is the response from the 7TV API for a channel's emotes.
type ChannelEmoteResponse struct {
	ID          string 	`json:"id"`
	Platform    string  `json:"platform"`
	Username    string  `json:"username"`
	DisplayName string  `json:"display_name"`
	// Unix timestamp
	LinkedAt      int64    `json:"linked_at"`
	EmoteCapacity int      `json:"emote_capacity"`
	EmoteSet      EmoteSet `json:"emote_set"`
}

// EmoteSet is a set of emotes
type EmoteSet struct {
	ID         primitive.ObjectID `json:"id"`
	Name       string             `json:"name"`
	Tags       []string           `json:"tags"`
	Immutable  bool               `json:"immutable"`
	Priviliged bool               `json:"priviliged"`
	Emotes     []Emote            `json:"emotes"`
	Capacity   int                `json:"capacity"`
	Owner      Owner              `json:"owner"`
}

// Emote is a single emote
type Emote struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Flags     int         `json:"flags"`
	Timestamp int64       `json:"timestamp"`
	ActorID   interface{} `json:"actor_id"`
	Data      Data        `json:"data"`
}

// Owner is the owner of an emoteset
type Owner struct {
	ID          string      `json:"id"`
	Username    string      `json:"username"`
	DisplayName string      `json:"display_name"`
	AvatarURL   string      `json:"avatar_url"`
	Roles       []string    `json:"roles"`
	Connections interface{} `json:"connections"`
}

// Files is a list of files which an emote has.
// It has information regarding the type, GIF, WEBP and AFIV.
// The "Host" struct which hosts Files, has a URL field which is the base URL for all files.
type Files struct {
	Name       string `json:"name"`
	StaticName string `json:"static_name"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Size       int    `json:"size"`
	Format     string `json:"format"`
}

// Host is the host of an emote
type Host struct {
	URL   string  `json:"url"`
	Files []Files `json:"files"`
}

// Data is metadata about an emote
type Data struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Flags     int    `json:"flags"`
	Lifecycle int    `json:"lifecycle"`
	Listed    bool   `json:"listed"`
	Animated  bool   `json:"animated"`
	Owner     Owner  `json:"owner"`
	Host      Host   `json:"host"`
}
