package ffz

// GlobalEmoteResponse is the response from the FFZ API for global emotes.
type GlobalEmoteResponse struct {
	DefaultSets []int          `json:"default_sets"`
	Sets        map[string]Set `json:"sets"`
}

// ChannelEmoteResponse is the response from the FFZ API for a channel's emotes.
type ChannelEmoteResponse struct {
	Room Room           `json:"room"`
	Sets map[string]Set `json:"sets"`
}

// Room is the room object in the FFZ API response.
type Room struct {
	ID       int `json:"_id"`
	// Why ffz :(
	TwitchID int `json:"twitch_id"`
}

// Owner is the owner of a emote
type Owner struct {
	ID          int    `json:"_id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

// Urls is a list of CDN URLs for an emote.
type Urls struct {
	Num1 string `json:"1"`
	Num2 string `json:"2"`
	Num4 string `json:"4"`
}

// Emoticon is a single emote.
type Emoticon struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Urls Urls   `json:"urls"`
}

// Set is a set of emotes.
//
// Typically a set is a representation of a twitch channel.
type Set struct {
	ID        int        `json:"id"`
	Type      int        `json:"_type"`
	Emoticons []Emoticon `json:"emoticons"`
}
