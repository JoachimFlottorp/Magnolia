package ffz

import (
	"context"
	"net/http"
	"strconv"
	"testing"

	"github.com/JoachimFlottorp/magnolia/external/client"
	"github.com/JoachimFlottorp/magnolia/external/emotes/models"
	"github.com/stretchr/testify/assert"
)

var testEmotes = []*models.Emote{
	{
		Name: "test",
		ImageList: models.ImageList{
			Small: "https://cdn.frankerfacez.com/emoticon/1/1",
			Big:   "https://cdn.frankerfacez.com/emoticon/1/4",
		},
	},
	{
		Name: "test2",
		ImageList: models.ImageList{
			Small: "https://cdn.frankerfacez.com/emoticon/2/1",
			Big:   "https://cdn.frankerfacez.com/emoticon/2/4",
		},
	},
}

func testClient(cb func(r *http.Request) (*http.Response, error)) *FFZ {
	return New(client.Options{
		Transport: client.Middleware(cb),
	})
}

func TestName(t *testing.T) {
	c := New(client.Options{})
	assert.Equal(t, "FFZ", c.Name().String())
}

func TestGetChannelEmotes(t *testing.T) {
	identifier := models.ChannelIdentifier{
		ID:   "123",
		Name: "test",
	}

	id, err := strconv.Atoi(identifier.ID)

	assert.NoError(t, err)
	
	input := ChannelEmoteResponse{
		Room: Room{
			ID:       123,
			TwitchID: id,
		},
		Sets: map[string]Set{
			"123": {
				ID: 123,
				Emoticons: []Emoticon{
					{
						ID:   1,
						Name: "test",
						Urls: Urls{
							Num1: "https://cdn.frankerfacez.com/emoticon/1/1",
							Num4: "https://cdn.frankerfacez.com/emoticon/1/4",
						},
					},
					{
						ID:   2,
						Name: "test2",
						Urls: Urls{
							Num1: "https://cdn.frankerfacez.com/emoticon/2/1",
							Num4: "https://cdn.frankerfacez.com/emoticon/2/4",
						},
					},
				},
			},
		},
	}

	ctx := context.Background()

	c := testClient(func(r *http.Request) (*http.Response, error) {
		assert.Equal(t, "https://api.frankerfacez.com/v1/room/test", r.URL.String())

		return client.JSONResponseTest(200, input)
	})

	eResp, err := c.GetChannelEmotes(ctx, identifier)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(eResp))

	for i, e := range eResp {
		assert.Equal(t, testEmotes[i].Name, e.Name)
		assert.Equal(t, testEmotes[i].ImageList, e.ImageList)
	}
}

func TestFetchGlobalEmotes(t *testing.T) {
	input := GlobalEmoteResponse{
		DefaultSets: []int{123},
		Sets: map[string]Set{
			"123": {
				ID: 123,
				Emoticons: []Emoticon{
					{
						ID:   1,
						Name: "test",
						Urls: Urls{
							Num1: "https://cdn.frankerfacez.com/emoticon/1/1",
							Num4: "https://cdn.frankerfacez.com/emoticon/1/4",
						},
					},
					{
						ID:   2,
						Name: "test2",
						Urls: Urls{
							Num1: "https://cdn.frankerfacez.com/emoticon/2/1",
							Num4: "https://cdn.frankerfacez.com/emoticon/2/4",
						},
					},
				},
			},
			"456": {
				ID: 456,
				Emoticons: []Emoticon{
					{
						ID:   3,
						Name: "test3",
						Urls: Urls{
							Num1: "https://cdn.frankerfacez.com/emoticon/3/1",
							Num4: "https://cdn.frankerfacez.com/emoticon/3/4",
						},
					},
					{
						ID:   4,
						Name: "test4",
						Urls: Urls{
							Num1: "https://cdn.frankerfacez.com/emoticon/4/1",
							Num4: "https://cdn.frankerfacez.com/emoticon/4/4",
						},
					},
				},
			},
		},
	}

	ctx := context.Background()

	c := testClient(func(r *http.Request) (*http.Response, error) {
		assert.Equal(t, "https://api.frankerfacez.com/v1/set/global", r.URL.String())

		return client.JSONResponseTest(200, input)
	})

	eResp, err := c.FetchGlobalEmotes(ctx)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(eResp))

	for i, e := range eResp {
		assert.Equal(t, testEmotes[i].Name, e.Name)
		assert.Equal(t, testEmotes[i].ImageList, e.ImageList)
	}
}

func TestGlobalCache(t *testing.T) {
	input := GlobalEmoteResponse{
		DefaultSets: []int{123},
		Sets: map[string]Set{
			"123": {
				ID: 123,
				Emoticons: []Emoticon{
					{
						ID:   1,
						Name: "test",
						Urls: Urls{
							Num1: "https://cdn.frankerfacez.com/emoticon/1/1",
							Num4: "https://cdn.frankerfacez.com/emoticon/1/4",
						},
					},
					{
						ID:   2,
						Name: "test2",
						Urls: Urls{
							Num1: "https://cdn.frankerfacez.com/emoticon/2/1",
							Num4: "https://cdn.frankerfacez.com/emoticon/2/4",
						},
					},
				},
			},
		},
	}

	ctx := context.Background()

	c := testClient(func(r *http.Request) (*http.Response, error) {
		assert.Equal(t, "https://api.frankerfacez.com/v1/set/global", r.URL.String())

		return client.JSONResponseTest(200, input)
	})

	err := c.UpdateGlobalCache(ctx)
	assert.NoError(t, err)

	eResp := c.GetGlobalEmotes()

	assert.Equal(t, 2, len(*eResp))

	for i, e := range *eResp {
		assert.Equal(t, testEmotes[i].Name, e.Name)
		assert.Equal(t, testEmotes[i].ImageList, e.ImageList)
	}
}

func TestNormalizeChannelResponse(t *testing.T) {
	testCase := &ChannelEmoteResponse{
		Sets: map[string]Set{
			"123": {
				Emoticons: []Emoticon{
					{
						ID:   1,
						Name: "test",
						Urls: Urls{
							Num1: "https://cdn.frankerfacez.com/emoticon/1/1",
							Num4: "https://cdn.frankerfacez.com/emoticon/1/4",
						},
					},
					{
						ID:   2,
						Name: "test2",
						Urls: Urls{
							Num1: "https://cdn.frankerfacez.com/emoticon/2/1",
							Num4: "https://cdn.frankerfacez.com/emoticon/2/4",
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
				Small: "https://cdn.frankerfacez.com/emoticon/1/1",
				Big:   "https://cdn.frankerfacez.com/emoticon/1/4",
			},
			Provider: models.ProviderTypeFFZ,
			Type:     models.EmoteTypeChannel,
		},
		{
			Name: "test2",
			ImageList: models.ImageList{
				Small: "https://cdn.frankerfacez.com/emoticon/2/1",
				Big:   "https://cdn.frankerfacez.com/emoticon/2/4",
			},
			Provider: models.ProviderTypeFFZ,
			Type:     models.EmoteTypeChannel,
		},
	}

	resp := fromChannelEmotesResponse(testCase)

	assert.Equal(t, len(expected), len(resp))

	for j, e := range resp {
		assert.Equal(t, expected[j].Name, e.Name)
		assert.Equal(t, expected[j].ImageList, e.ImageList)
		assert.Equal(t, expected[j].Provider, e.Provider)
		assert.Equal(t, expected[j].Type, e.Type)
	}
}

func TestNormalizeGlobalResponse(t *testing.T) {

}

func TestContains(t *testing.T) {
	type testData struct {
		Slice    []int
		Value    int
		Expected bool
	}

	testCases := []testData{
		{
			Slice:    []int{1, 2, 3},
			Value:    1,
			Expected: true,
		},
		{
			Slice:    []int{1, 2, 3},
			Value:    4,
			Expected: false,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.Expected, contains(tc.Slice, tc.Value))
	}
}
