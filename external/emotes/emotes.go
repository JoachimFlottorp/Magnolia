package emotes

import (
	"context"

	"github.com/JoachimFlottorp/GoCommon/cron"
	"github.com/JoachimFlottorp/magnolia/external/client"
	"github.com/JoachimFlottorp/magnolia/external/emotes/bttv"
	"github.com/JoachimFlottorp/magnolia/external/emotes/ffz"
	"github.com/JoachimFlottorp/magnolia/external/emotes/models"
	"github.com/JoachimFlottorp/magnolia/external/emotes/seventv"
	"go.uber.org/zap"
)

/*
	TODO: Do some improvements, idk.
*/

type EmoteProvider interface {
	// Name returns the name of the provider
	Name() models.ProviderType
	// GetChannelEmotes returns the emotes for a channel
	GetChannelEmotes(context.Context, models.ChannelIdentifier) ([]*models.Emote, error)
	// GetGlobalEmotes returns the global emotes
	GetGlobalEmotes() *[]*models.Emote
	// FetchGlobalEmotes fetches the global emotes and stores them in the cache
	FetchGlobalEmotes(context.Context) ([]*models.Emote, error)
	// UpdateGlobalCache updates the global cache
	UpdateGlobalCache(context.Context) error
}

type EmoteProviderList []EmoteProvider

func (e *EmoteProviderList) GetProvider(name models.ProviderType) EmoteProvider {
	for _, provider := range *e {
		if provider.Name() == name {
			return provider
		}
	}

	return nil
}

var Providers EmoteProviderList
var CronTab = cron.NewManager(context.Background(), false)

func init() {
	Providers = append(Providers, seventv.New(client.Options{}))
	Providers = append(Providers, bttv.New(client.Options{}))
	Providers = append(Providers, ffz.New(client.Options{}))

	CronTab.Add(cron.CronOptions{
		Name:   "ffz:update_global_emotes",
		Spec:   "0 0 */24 * * *",
		RunNow: true,
		Cmd: func() {
			ffz := Providers.GetProvider(models.ProviderTypeFFZ)
			if ffz == nil {
				zap.S().Error("failed to get ffz provider")
				return
			}

			err := ffz.UpdateGlobalCache(context.Background())

			if err != nil {
				zap.S().Error("failed to update global emotes", zap.Error(err))
				return
			}
		},
	})

	CronTab.Add(cron.CronOptions{
		Name:   "bttv:update_global_emotes",
		Spec:   "0 0 */24 * * *",
		RunNow: true,
		Cmd: func() {
			bttv := Providers.GetProvider(models.ProviderTypeBTTV)
			if bttv == nil {
				zap.S().Error("failed to get bttv provider")
				return
			}

			err := bttv.UpdateGlobalCache(context.Background())

			if err != nil {
				zap.S().Error("failed to update global emotes", zap.Error(err))
				return
			}

		},
	})

	CronTab.Add(cron.CronOptions{
		Name:   "7tv:update_global_emotes",
		Spec:   "0 0 */24 * * *",
		RunNow: true,
		Cmd: func() {
			seventv := Providers.GetProvider(models.ProviderType7TV)
			if seventv == nil {
				zap.S().Error("failed to get 7tv provider")
				return
			}

			err := seventv.UpdateGlobalCache(context.Background())

			if err != nil {
				zap.S().Error("failed to update global emotes", zap.Error(err))
				return
			}
		},
	})

	CronTab.Start()
}

func GetEmotes(ctx context.Context, id models.ChannelIdentifier) (*models.EmoteList, error) {
	var emotes []*models.Emote

	for _, provider := range Providers {
		emote, err := provider.GetChannelEmotes(ctx, id)

		if err != nil {
			return nil, err
		}

		emotes = append(emotes, emote...)
	}

	return FilterEmotes(emotes), nil
}

// TODO: Could be improved
func FilterEmotes(emotes []*models.Emote) *models.EmoteList {
	var filteredEmotes []*models.Emote
	var dupl = make(map[string]models.EmoteHierarchyFlag)

	for _, emote := range emotes {
		if emote == nil {
			continue
		}

		if emote.Hierarchy() == 0 {
			continue
		}

		if _, ok := dupl[emote.Name]; ok {
			if dupl[emote.Name] > emote.Hierarchy() {
				continue
			}

			dupl[emote.Name] = emote.Hierarchy()

			continue
		}

		dupl[emote.Name] = emote.Hierarchy()

		filteredEmotes = append(filteredEmotes, emote)
	}

	return &models.EmoteList{
		Emotes: filteredEmotes,
	}
}
