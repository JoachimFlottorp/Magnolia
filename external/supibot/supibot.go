package supibot

import (
	"context"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

const BASE_URL = "https://supinic.com/api/"

type OuterResponse[T any] struct {
	StatusCode int         `json:"statusCode"`
	Timestamp  int         `json:"timestamp"`
	Data       T           `json:"data"`
	Error      interface{} `json:"error"`
}

type BotList_t struct {
	Bots []struct {
		ID                int       `json:"id"`
		Name              string    `json:"name"`
		Prefix            string    `json:"prefix"`
		HasPrefixSpace    bool      `json:"hasPrefixSpace"`
		AuthorID          int       `json:"authorID"`
		AuthorName        string    `json:"authorName"`
		Language          string    `json:"language"`
		Description       string    `json:"description"`
		Level             int       `json:"level"`
		LastSeen          string	`json:"lastSeen"`
		LastSeenTimestamp int       `json:"lastSeenTimestamp"`
	} `json:"bots"`
}

func GetBotList(ctx context.Context, client *http.Client) (*OuterResponse[BotList_t], error) {
	req, err := http.NewRequestWithContext(ctx, "GET", BASE_URL+"bot-program/bot/list", nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var data OuterResponse[BotList_t]
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		zap.S().Errorf("[Supibot] Error bot/list: %s", data)
		return nil, err
	}

	return &data, nil
}
