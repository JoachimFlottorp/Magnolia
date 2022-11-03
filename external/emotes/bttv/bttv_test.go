package bttv

import (
	"context"
	"net/http"
	"testing"

	"github.com/JoachimFlottorp/magnolia/external/client"
	"github.com/JoachimFlottorp/magnolia/external/emotes/models"
	"github.com/stretchr/testify/assert"
)

func testClient(cb func(r *http.Request) (*http.Response, error)) *BTTV {
	return New(client.Options{
		Transport: client.Middleware(cb),
	})
}

func TestName(t *testing.T) {
	b := New(client.Options{})
	assert.Equal(t, b.Name().String(), "BTTV")
}

func TestGetChannelEmotes(t *testing.T) {
	identifier := models.ChannelIdentifier{
		ID:   "123",
		Name: "test",
	}

	input := ChannelEmoteResponse{
		ChannelEmotes: []OwnedEmote{
			{
				ID:        "1",
				Code:      "test",
				ImageType: "png",
			},
			{
				ID:        "2",
				Code:      "test2",
				ImageType: "png",
			},
		},
		SharedEmotes: []SharedEmotes{
			{
				ID:        "3",
				Code:      "test3",
				ImageType: "png",
			},
			{
				ID:        "4",
				Code:      "test4",
				ImageType: "png",
			},
		},
	}

	expected := []*models.Emote{
		{
			Name: "test",
			ImageList: models.ImageList{
				Small: "https://cdn.betterttv.net/emote/1/1x",
				Big:   "https://cdn.betterttv.net/emote/1/3x",
			},
		},
		{
			Name: "test2",
			ImageList: models.ImageList{
				Small: "https://cdn.betterttv.net/emote/2/1x",
				Big:   "https://cdn.betterttv.net/emote/2/3x",
			},
		},
		{
			Name: "test3",
			ImageList: models.ImageList{
				Small: "https://cdn.betterttv.net/emote/3/1x",
				Big:   "https://cdn.betterttv.net/emote/3/3x",
			},
		},
		{
			Name: "test4",
			ImageList: models.ImageList{
				Small: "https://cdn.betterttv.net/emote/4/1x",
				Big:   "https://cdn.betterttv.net/emote/4/3x",
			},
		},
	}

	ctx := context.Background()

	b := testClient(func(r *http.Request) (*http.Response, error) {
		return client.JSONResponseTest(200, input)
	})

	emotes, err := b.GetChannelEmotes(ctx, identifier)
	assert.NoError(t, err)

	assert.Equal(t, len(emotes), len(expected))

	for i, e := range emotes {
		assert.Equal(t, e.Name, expected[i].Name)
		assert.Equal(t, e.ImageList.Small, expected[i].ImageList.Small)
		assert.Equal(t, e.ImageList.Big, expected[i].ImageList.Big)
	}
}

func TestGetGlobalEmotes(t *testing.T) {
	input := GlobalEmoteResponse{
		{
			ID:        "1",
			Code:      "test",
			ImageType: "png",
		},
		{
			ID:        "2",
			Code:      "test2",
			ImageType: "png",
		},
	}

	expected := []*models.Emote{
		{
			Name: "test",
			ImageList: models.ImageList{
				Small: "https://cdn.betterttv.net/emote/1/1x",
				Big:   "https://cdn.betterttv.net/emote/1/3x",
			},
		},
		{
			Name: "test2",
			ImageList: models.ImageList{
				Small: "https://cdn.betterttv.net/emote/2/1x",
				Big:   "https://cdn.betterttv.net/emote/2/3x",
			},
		},
	}

	ctx := context.Background()

	b := testClient(func(r *http.Request) (*http.Response, error) {
		return client.JSONResponseTest(200, input)
	})

	emotes, err := b.FetchGlobalEmotes(ctx)
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
		{
			ID:        "1",
			Code:      "test",
			ImageType: "png",
		},
		{
			ID:        "2",
			Code:      "test2",
			ImageType: "png",
		},
	}

	expected := []*models.Emote{
		{
			Name: "test",
			ImageList: models.ImageList{
				Small: "https://cdn.betterttv.net/emote/1/1x",
				Big:   "https://cdn.betterttv.net/emote/1/3x",
			},
		},
		{
			Name: "test2",
			ImageList: models.ImageList{
				Small: "https://cdn.betterttv.net/emote/2/1x",
				Big:   "https://cdn.betterttv.net/emote/2/3x",
			},
		},
	}

	ctx := context.Background()

	b := testClient(func(r *http.Request) (*http.Response, error) {
		return client.JSONResponseTest(200, input)
	})

	err := b.UpdateGlobalCache(ctx)
	assert.NoError(t, err)

	emotes := b.GetGlobalEmotes()

	assert.Equal(t, len(*emotes), len(expected))

	for i, e := range *emotes {
		assert.Equal(t, e.Name, expected[i].Name)
		assert.Equal(t, e.ImageList.Small, expected[i].ImageList.Small)
		assert.Equal(t, e.ImageList.Big, expected[i].ImageList.Big)
	}
}
