package bttv

// GlobalEmoteResponse is the response from the BTTV API for global emotes.
type GlobalEmoteResponse []OwnedEmote

// ChannelEmoteResponse is the response from the BTTV API for a channel's emotes.
type ChannelEmoteResponse struct {
	ID            string         `json:"id"`
	Bots          []string       `json:"bots"`
	Avatar        string         `json:"avatar"`
	ChannelEmotes []OwnedEmote   `json:"channelEmotes"`
	SharedEmotes  []SharedEmotes `json:"sharedEmotes"`
}

// OwnedEmote are emotes that have been uploaded by the channel owner.
// Or global emotes that have been uploaded by the BTTV team.
type OwnedEmote struct {
	ID        string `json:"id"`
	Code      string `json:"code"`
	ImageType string `json:"imageType"`
	UserID    string `json:"userId"`
}

// User a user that has uploaded emotes to BTTV.
type User struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	ProviderID  string `json:"providerId"`
}

// SharedEmotes are emotes that have been uploaded by other channels.
type SharedEmotes struct {
	ID        string `json:"id"`
	Code      string `json:"code"`
	ImageType string `json:"imageType"`
	User      User   `json:"user"`
}
