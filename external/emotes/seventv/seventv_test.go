package seventv

import (
	"context"
	"net/http"
	"testing"

	"github.com/JoachimFlottorp/magnolia/external/client"
	"github.com/JoachimFlottorp/magnolia/external/emotes/models"
	"github.com/stretchr/testify/assert"
)

func testClient(cb func(r *http.Request) (*http.Response, error)) *SevenTV {
	return New(client.Options{
		Transport: client.Middleware(cb),
	})
}

func TestName(t *testing.T) {
	s := New(client.Options{})
	assert.Equal(t, s.Name().String(), "7TV")
}

func TestGetChannelEmotes(t *testing.T) {
	identifier := models.ChannelIdentifier{
		ID:   "123",
		Name: "test",
	}

	input := ChannelEmoteResponse{
		EmoteSet: EmoteSet{
			Emotes: []Emote{
				{
					ID:   "1",
					Name: "test",
					Data: Data{
						Host: Host{
							URL: "//cdn.7tv.app/emote/1",
						},
					},
				},
				{
					ID:   "2",
					Name: "test2",
					Data: Data{
						Host: Host{
							URL: "//cdn.7tv.app/emote/2",
						},
					},
				},
			},
		},
	}

	expected := []*models.Emote{
		{
			Name: "test",
			ImageList: models.ImageList{
				Small: "https://cdn.7tv.app/emote/1/1x.webp",
				Big:   "https://cdn.7tv.app/emote/1/4x.webp",
			},
		},
		{
			Name: "test2",
			ImageList: models.ImageList{
				Small: "https://cdn.7tv.app/emote/2/1x.webp",
				Big:   "https://cdn.7tv.app/emote/2/4x.webp",
			},
		},
	}

	ctx := context.Background()

	s := testClient(func(r *http.Request) (*http.Response, error) {
		return client.JSONResponseTest(200, input)
	})

	emotes, err := s.GetChannelEmotes(ctx, identifier)
	assert.NoError(t, err)

	assert.Equal(t, len(emotes), len(expected))

	for i, e := range emotes {
		assert.Equal(t, e.Name, expected[i].Name)
		assert.Equal(t, e.ImageList.Small, expected[i].ImageList.Small)
		assert.Equal(t, e.ImageList.Big, expected[i].ImageList.Big)
	}
}

func TestFetchGlobalEmotes(t *testing.T) {
	input := GlobalEmoteResponse{
		EmoteSet: EmoteSet{
			Emotes: []Emote{
				{
					ID:   "1",
					Name: "test",
					Data: Data{
						Host: Host{
							URL: "//cdn.7tv.app/emote/1",
						},
					},
				},
				{
					ID:   "2",
					Name: "test2",
					Data: Data{
						Host: Host{
							URL: "//cdn.7tv.app/emote/2",
						},
					},
				},
			},
		},
	}

	expected := []*models.Emote{
		{
			Name: "test",
			ImageList: models.ImageList{
				Small: "https://cdn.7tv.app/emote/1/1x.webp",
				Big:   "https://cdn.7tv.app/emote/1/4x.webp",
			},
		},
		{
			Name: "test2",
			ImageList: models.ImageList{
				Small: "https://cdn.7tv.app/emote/2/1x.webp",
				Big:   "https://cdn.7tv.app/emote/2/4x.webp",
			},
		},
	}

	ctx := context.Background()

	s := testClient(func(r *http.Request) (*http.Response, error) {
		return client.JSONResponseTest(200, input)
	})

	emotes, err := s.FetchGlobalEmotes(ctx)
	assert.NoError(t, err)

	assert.Equal(t, len(emotes), len(expected))

	for i, e := range emotes {
		assert.Equal(t, e.Name, expected[i].Name)
		assert.Equal(t, e.ImageList.Small, expected[i].ImageList.Small)
		assert.Equal(t, e.ImageList.Big, expected[i].ImageList.Big)
	}
}

func TestGlobalCache(t *testing.T) {
	input := GlobalEmoteResponse{
		EmoteSet: EmoteSet{
			Emotes: []Emote{
				{
					ID:   "1",
					Name: "test",
					Data: Data{
						Host: Host{
							URL: "//cdn.7tv.app/emote/1",
						},
					},
				},
				{
					ID:   "2",
					Name: "test2",
					Data: Data{
						Host: Host{
							URL: "//cdn.7tv.app/emote/2",
						},
					},
				},
			},
		},
	}

	expected := []*models.Emote{
		{
			Name: "test",
			ImageList: models.ImageList{
				Small: "https://cdn.7tv.app/emote/1/1x.webp",
				Big:   "https://cdn.7tv.app/emote/1/4x.webp",
			},
		},
		{
			Name: "test2",
			ImageList: models.ImageList{
				Small: "https://cdn.7tv.app/emote/2/1x.webp",
				Big:   "https://cdn.7tv.app/emote/2/4x.webp",
			},
		},
	}

	ctx := context.Background()

	s := testClient(func(r *http.Request) (*http.Response, error) {
		return client.JSONResponseTest(200, input)
	})

	err := s.UpdateGlobalCache(ctx)

	assert.NoError(t, err)

	emotes := s.GetGlobalEmotes()

	assert.Equal(t, len(*emotes), len(expected))

	for i, e := range *emotes {
		assert.Equal(t, e.Name, expected[i].Name)
		assert.Equal(t, e.ImageList.Small, expected[i].ImageList.Small)
		assert.Equal(t, e.ImageList.Big, expected[i].ImageList.Big)
	}
}
