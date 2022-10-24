package ivr

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"
)

const BASE_URL = "https://api.ivr.fi/v2/"

type ResolveUser_t struct {
	Banned           bool          `json:"banned"`
	DisplayName      string        `json:"displayName"`
	Login            string        `json:"login"`
	ID               string        `json:"id"`
	Bio              string        `json:"bio"`
	Follows          int           `json:"follows"`
	Followers        int           `json:"followers"`
	ProfileViewCount int           `json:"profileViewCount"`
	ChatColor        string        `json:"chatColor"`
	Logo             string        `json:"logo"`
	Banner           interface{}   `json:"banner"`
	VerifiedBot      bool          `json:"verifiedBot"`
	CreatedAt        time.Time     `json:"createdAt"`
	UpdatedAt        time.Time     `json:"updatedAt"`
	EmotePrefix      string        `json:"emotePrefix"`
	Roles            Roles         `json:"roles"`
	Badges           []Badges      `json:"badges"`
	ChatSettings     ChatSettings  `json:"chatSettings"`
	Stream           interface{}   `json:"stream"`
	LastBroadcast    LastBroadcast `json:"lastBroadcast"`
	Panels           []Panels      `json:"panels"`
}

type Roles struct {
	IsAffiliate bool        `json:"isAffiliate"`
	IsPartner   bool        `json:"isPartner"`
	IsStaff     interface{} `json:"isStaff"`
}

type Badges struct {
	SetID       string `json:"setID"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

type ChatSettings struct {
	ChatDelayMs                  int           `json:"chatDelayMs"`
	FollowersOnlyDurationMinutes interface{}   `json:"followersOnlyDurationMinutes"`
	SlowModeDurationSeconds      interface{}   `json:"slowModeDurationSeconds"`
	BlockLinks                   bool          `json:"blockLinks"`
	IsSubscribersOnlyModeEnabled bool          `json:"isSubscribersOnlyModeEnabled"`
	IsEmoteOnlyModeEnabled       bool          `json:"isEmoteOnlyModeEnabled"`
	IsFastSubsModeEnabled        bool          `json:"isFastSubsModeEnabled"`
	IsUniqueChatModeEnabled      bool          `json:"isUniqueChatModeEnabled"`
	RequireVerifiedAccount       bool          `json:"requireVerifiedAccount"`
	Rules                        []interface{} `json:"rules"`
}

type LastBroadcast struct {
	StartedAt time.Time `json:"startedAt"`
	Title     string    `json:"title"`
}

type Panels struct {
	ID string `json:"id"`
}

func ResolveUsernames(ctx context.Context, usernames []string) ([]*ResolveUser_t, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", BASE_URL+"twitch/user", nil)
	if err != nil { return nil, err }

	q := req.URL.Query()
	for _, username := range usernames {
		q.Add("login", username)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil { return nil, err }

	defer resp.Body.Close()
	
	var data []*ResolveUser_t
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil { return nil, err }

	if resp.StatusCode != http.StatusOK {
		zap.S().Errorf("[IVR] Error twitch/user: %s", data)
		return nil, err
	}

	return data, nil
}