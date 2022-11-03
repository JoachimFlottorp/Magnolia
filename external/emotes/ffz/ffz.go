package ffz

import (
	"context"
	"fmt"
	"net/http"

	"github.com/JoachimFlottorp/magnolia/external/client"
	"github.com/JoachimFlottorp/magnolia/external/emotes/models"
)

const (
	API_CHANNEL_EMOTES = "https://api.frankerfacez.com/v1/room/%s"
	API_GLOBAL_EMOTES  = "https://api.frankerfacez.com/v1/set/global"
	API_CDN            = "https://cdn.frankerfacez.com/emoticon/%s/%s"
)

type FFZ struct {
	*client.Client

	GlobalCache []*models.Emote
}

func New(opts client.Options) *FFZ {
	if opts.Transport == nil {
		opts.Transport = http.DefaultTransport
	}

	return &FFZ{
		GlobalCache: []*models.Emote{},

		Client: client.NewClient(opts),
	}
}

func (b *FFZ) Name() models.ProviderType {
	return models.ProviderTypeFFZ
}

func (b *FFZ) GetChannelEmotes(ctx context.Context, identifier models.ChannelIdentifier) ([]*models.Emote, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(API_CHANNEL_EMOTES, identifier.Name), nil)

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

	return fromChannelEmotesResponse(body), nil
}

func (b *FFZ) FetchGlobalEmotes(ctx context.Context) ([]*models.Emote, error) {
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

	return fromGlobalEmotesResponse(body), nil
}

func (b *FFZ) UpdateGlobalCache(ctx context.Context) error {
	emotes, err := b.FetchGlobalEmotes(ctx)

	if err != nil {
		return err
	}

	b.GlobalCache = emotes

	return nil
}

func (b *FFZ) GetGlobalEmotes() *[]*models.Emote {
	return &b.GlobalCache
}

func constructImages(id string) models.ImageList {
	return models.ImageList{
		Small: fmt.Sprintf(API_CDN, id, "1"),
		Big:   fmt.Sprintf(API_CDN, id, "4"),
	}
}

func fromChannelEmotesResponse(channelEmotesResponse *ChannelEmoteResponse) []*models.Emote {
	var emotes []*models.Emote

	for _, set := range channelEmotesResponse.Sets {
		for _, emote := range set.Emoticons {
			emotes = append(emotes, &models.Emote{
				Name:      emote.Name,
				Type:      models.EmoteTypeChannel,
				Provider:  models.ProviderTypeFFZ,
				ImageList: constructImages(fmt.Sprintf("%d", emote.ID)),
			})
		}
	}

	return emotes
}

func fromGlobalEmotesResponse(globalEmotesResponse *GlobalEmoteResponse) []*models.Emote {
	var emotes []*models.Emote

	defaultSets := globalEmotesResponse.DefaultSets

	for _, set := range globalEmotesResponse.Sets {
		if !contains(defaultSets, set.ID) {
			continue
		}

		for _, emote := range set.Emoticons {
			emotes = append(emotes, &models.Emote{
				Name:      emote.Name,
				Type:      models.EmoteTypeGlobal,
				Provider:  models.ProviderTypeFFZ,
				ImageList: constructImages(fmt.Sprintf("%d", emote.ID)),
			})
		}
	}

	return emotes
}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
