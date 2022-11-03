package bttv

import (
	"context"
	"fmt"
	"net/http"

	"github.com/JoachimFlottorp/magnolia/external/client"
	"github.com/JoachimFlottorp/magnolia/external/emotes/models"
)

const (
	API_CHANNEL_EMOTES = "https://api.betterttv.net/3/cached/users/twitch/%s"
	API_GLOBAL_EMOTES  = "https://api.betterttv.net/3/cached/emotes/global"
	API_CDN            = "https://cdn.betterttv.net/emote/%s/%s"
)

type BTTV struct {
	*client.Client

	GlobalCache []*models.Emote
}

func New(opts client.Options) *BTTV {
	if opts.Transport == nil {
		opts.Transport = http.DefaultTransport
	}

	return &BTTV{
		Client:      client.NewClient(opts),
		GlobalCache: []*models.Emote{},
	}
}

func (b *BTTV) Name() models.ProviderType {
	return models.ProviderTypeBTTV
}

func (b *BTTV) GetChannelEmotes(ctx context.Context, identifier models.ChannelIdentifier) ([]*models.Emote, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(API_CHANNEL_EMOTES, identifier.ID), nil)

	if err != nil {
		return nil, err
	}

	resp, err := b.Do(req)

	if err != nil {
		return nil, err
	}

	body, err := client.DoJSON[ChannelEmoteResponse](resp)
	if err != nil {
		return nil, err
	}

	return fromChannelResponse(body), nil
}

func (b *BTTV) FetchGlobalEmotes(ctx context.Context) ([]*models.Emote, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, API_GLOBAL_EMOTES, nil)

	if err != nil {
		return nil, err
	}

	resp, err := b.Do(req)

	if err != nil {
		return nil, err
	}

	body, err := client.DoJSON[GlobalEmoteResponse](resp)
	if err != nil {
		return nil, err
	}

	return fromGlobalResponse(body), nil
}

func (b *BTTV) UpdateGlobalCache(ctx context.Context) error {
	emotes, err := b.FetchGlobalEmotes(ctx)

	if err != nil {
		return err
	}

	b.GlobalCache = emotes

	return nil
}

func (b *BTTV) GetGlobalEmotes() *[]*models.Emote {
	return &b.GlobalCache
}

func constructImages(id string) models.ImageList {
	return models.ImageList{
		Small: fmt.Sprintf(API_CDN, id, "1x"),
		Big:   fmt.Sprintf(API_CDN, id, "3x"),
	}
}

func fromGlobalResponse(response *GlobalEmoteResponse) []*models.Emote {
	var emotes []*models.Emote

	for _, emote := range *response {
		emotes = append(emotes, &models.Emote{
			Name:      emote.Code,
			Provider:  models.ProviderTypeBTTV,
			Type:      models.EmoteTypeGlobal,
			ImageList: constructImages(emote.ID),
		})
	}

	return emotes
}

func fromChannelResponse(response *ChannelEmoteResponse) []*models.Emote {
	var emotes []*models.Emote

	for _, emote := range response.ChannelEmotes {
		emotes = append(emotes, &models.Emote{
			Name:      emote.Code,
			Provider:  models.ProviderTypeBTTV,
			Type:      models.EmoteTypeChannel,
			ImageList: constructImages(emote.ID),
		})
	}

	for _, emote := range response.SharedEmotes {
		emotes = append(emotes, &models.Emote{
			Name:      emote.Code,
			Provider:  models.ProviderTypeBTTV,
			Type:      models.EmoteTypeChannel,
			ImageList: constructImages(emote.ID),
		})
	}

	return emotes
}
