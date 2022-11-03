package emotes

import (
	"testing"

	"github.com/JoachimFlottorp/magnolia/external/emotes/models"
	"github.com/stretchr/testify/assert"
)

func TestCanFilterEmotes(t *testing.T) {
	emotes := []*models.Emote{
		{
			Provider: models.ProviderTypeFFZ,
			Name:     "test",
			Type:     models.EmoteTypeChannel,
		},
		{
			Provider: models.ProviderTypeBTTV,
			Name:     "test",
			Type:     models.EmoteTypeChannel,
		},

		{
			Provider: models.ProviderType7TV,
			Name:     "test2",
			Type:     models.EmoteTypeChannel,
		},
		nil,
		nil,
		{
			Provider: models.ProviderType(""),
			Name:     "test3",
			Type:     models.EmoteTypeChannel,
		},
		{
			Provider: models.ProviderType7TV,
			Name:     "test2",
			Type:     models.EmoteTypeChannel,
		},
		{
			Provider: models.ProviderTypeFFZ,
			Name:     "test",
			Type:     models.EmoteTypeChannel,
		},
	}

	expected := models.EmoteList{Emotes: []*models.Emote{
		{
			Provider: models.ProviderTypeFFZ,
			Name:     "test",
			Type:     models.EmoteTypeChannel,
		},
		{
			Provider: models.ProviderType7TV,
			Name:     "test2",
			Type:     models.EmoteTypeChannel,
		},
	}}

	filtered := FilterEmotes(emotes)

	assert.Equal(t, len(expected.Emotes), len(filtered.Emotes))

	assert.False(t, isDuplicates(filtered.Emotes))
}

func isDuplicates(emotes []*models.Emote) bool {
	seen := make(map[string]bool)

	for _, emote := range emotes {
		if seen[emote.Name] {
			return true
		}

		seen[emote.Name] = true
	}

	return false
}
