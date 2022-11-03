package seventv

import (
	"context"
	"fmt"
	"net/http"

	"github.com/JoachimFlottorp/magnolia/external/client"
	"github.com/JoachimFlottorp/magnolia/external/emotes/models"
	"go.uber.org/zap"
)

const (
	API_V3 = "https://7tv.io/v3/"

	API_GET_GLOBAL_EMOTES  = API_V3 + "emote-sets/global"
	API_GET_CHANNEL_EMOTES = API_V3 + "users/twitch/%s"
)

type SevenTV struct {
	*client.Client

	GlobalCache []*models.Emote
}

func New(opts client.Options) *SevenTV {
	if opts.Transport == nil {
		opts.Transport = http.DefaultTransport
	}

	return &SevenTV{
		Client:      client.NewClient(opts),
		GlobalCache: []*models.Emote{},
	}
}

func (s *SevenTV) Name() models.ProviderType {
	return models.ProviderType7TV
}

func (s *SevenTV) GetChannelEmotes(ctx context.Context, identifier models.ChannelIdentifier) ([]*models.Emote, error) {
	url := fmt.Sprintf(API_GET_CHANNEL_EMOTES, identifier.ID)
	zap.S().Debug(url)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		return nil, err
	}

	resp, err := s.Do(req)

	if err != nil {
		return nil, err
	}

	body, err := client.DoJSON[ChannelEmoteResponse](resp)

	if err != nil {
		return nil, err
	}

	return fromChannelResponse(body), nil
}

func (s *SevenTV) FetchGlobalEmotes(ctx context.Context) ([]*models.Emote, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, API_GET_GLOBAL_EMOTES, nil)

	if err != nil {
		return nil, err
	}

	resp, err := s.Do(req)

	if err != nil {
		return nil, err
	}

	body, err := client.DoJSON[GlobalEmoteResponse](resp)

	if err != nil {
		return nil, err
	}

	return fromGlobalResponse(body), nil
}

func (s *SevenTV) UpdateGlobalCache(ctx context.Context) error {
	emotes, err := s.FetchGlobalEmotes(ctx)

	if err != nil {
		return err
	}

	s.GlobalCache = emotes

	return nil
}

func (s *SevenTV) GetGlobalEmotes() *[]*models.Emote {
	return &s.GlobalCache
}

func constructImages(host Host) models.ImageList {
	url := "https:" + host.URL

	return models.ImageList{
		Big:   url + "/4x.webp",
		Small: url + "/1x.webp",
	}
}

func fromGlobalResponse(response *GlobalEmoteResponse) []*models.Emote {
	var emotes []*models.Emote

	for _, emote := range response.Emotes {
		emotes = append(emotes, &models.Emote{
			Name:      emote.Name,
			Provider:  models.ProviderType7TV,
			Type:      models.EmoteTypeGlobal,
			ImageList: constructImages(emote.Data.Host),
		})
	}

	return emotes
}

func fromChannelResponse(response *ChannelEmoteResponse) []*models.Emote {
	var emotes []*models.Emote

	for _, emote := range response.EmoteSet.Emotes {
		emotes = append(emotes, &models.Emote{
			Name:      emote.Name,
			Provider:  models.ProviderType7TV,
			Type:      models.EmoteTypeChannel,
			ImageList: constructImages(emote.Data.Host),
		})
	}

	return emotes
}
