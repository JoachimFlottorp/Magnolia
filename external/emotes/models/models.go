package models

import (
	"context"
	"fmt"

	"github.com/JoachimFlottorp/magnolia/internal/redis"
	pb "github.com/JoachimFlottorp/magnolia/protobuf"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type EmoteType string
type ProviderType string
type EmoteHierarchyFlag int32

const (
	EmoteTypeGlobal  = EmoteType("global")
	EmoteTypeChannel = EmoteType("channel")

	ProviderType7TV  = ProviderType("7TV")
	ProviderTypeBTTV = ProviderType("BTTV")
	ProviderTypeFFZ  = ProviderType("FFZ")

	EmoteHierarchyFFZ  = EmoteHierarchyFlag(1 << 2)
	EmoteHierarchyBTTV = EmoteHierarchyFlag(1 << 4)
	EmoteHierarchy7TV  = EmoteHierarchyFlag(1 << 8)
)

func (e EmoteType) String() string {
	return string(e)
}

func (e EmoteType) ToProto() pb.EmoteType {
	switch e {
	case EmoteTypeGlobal:
		return pb.EmoteType_GLOBAL
	case EmoteTypeChannel:
		return pb.EmoteType_CHANNEL
	}

	return pb.EmoteType_GLOBAL
}

func (p ProviderType) String() string {
	return string(p)
}

func (p ProviderType) ToProto() pb.ProviderType {
	switch p {
	case ProviderType7TV:
		return pb.ProviderType_SEVENTV
	case ProviderTypeBTTV:
		return pb.ProviderType_BTTV
	case ProviderTypeFFZ:
		return pb.ProviderType_FFZ
	}

	return pb.ProviderType_UNKNOWN
}

type ImageList struct {
	Small string
	Big   string
}

func (i *ImageList) ToProto() *pb.ImageList {
	return &pb.ImageList{
		Small: i.Small,
		Big:   i.Big,
	}
}

type ChannelIdentifier struct {
	ID   string
	Name string
}

type Emote struct {
	Name      string
	Provider  ProviderType
	Type      EmoteType
	ImageList ImageList
}

func (e *Emote) ToProto() *pb.Emote {
	return &pb.Emote{
		Name:      e.Name,
		Provider:  e.Provider.ToProto(),
		Type:      e.Type.ToProto(),
		ImageList: e.ImageList.ToProto(),
	}
}

func (e *Emote) Hierarchy() EmoteHierarchyFlag {
	switch e.Provider {
	case ProviderTypeFFZ:
		return EmoteHierarchyFFZ
	case ProviderTypeBTTV:
		return EmoteHierarchyBTTV
	case ProviderType7TV:
		return EmoteHierarchy7TV
	default:
		return 0
	}
}

func (e *Emote) IsGlobal() bool {
	return e.Type == EmoteTypeGlobal
}

func (e *Emote) IsChannel() bool {
	return e.Type == EmoteTypeChannel
}

type EmoteList struct {
	Emotes []*Emote
}

func (el *EmoteList) ToProto() *pb.EmoteList {
	emotes := make([]*pb.Emote, len(el.Emotes))
	for i, e := range el.Emotes {
		emotes[i] = e.ToProto()
	}

	return &pb.EmoteList{
		Emotes: emotes,
	}
}

func (el *EmoteList) Save(ctx context.Context, i redis.Instance, channel string) error {
	key := fmt.Sprintf("twitch:%s:emotes", channel)

	zap.S().Debugw("", "channel", channel, "length", len(el.Emotes))
	
	data, err := proto.Marshal(el.ToProto())
	if err != nil {
		return err
	}

	return i.Set(ctx, key, string(data))
}

func (el *EmoteList) Load(ctx context.Context, i redis.Instance, channel string) error {
	key := fmt.Sprintf("twitch:%s:emotes", channel)

	data, err := i.Get(ctx, key)
	if err != nil {
		return err
	}

	emote := &pb.EmoteList{}
	if err := proto.Unmarshal([]byte(data), emote); err != nil {
		return err
	}

	el.Emotes = make([]*Emote, len(emote.Emotes))

	for i, e := range emote.Emotes {
		el.Emotes[i] = &Emote{
			Name:     e.Name,
			Provider: ProviderType(e.Provider.String()),
			Type:     EmoteType(e.Type.String()),
			ImageList: ImageList{
				Small: e.ImageList.Small,
				Big:   e.ImageList.Big,
			},
		}
	}

	return nil
}

func (el *EmoteList) Delete(ctx context.Context, i redis.Instance, channel string) error {
	key := fmt.Sprintf("twitch:%s:emotes", channel)

	return i.Del(ctx, key)
}

func (el *EmoteList) SaveGlobal(ctx context.Context, i redis.Instance) error {
	key := "twitch:global:emotes"

	data, err := proto.Marshal(el.ToProto())
	if err != nil {
		return err
	}

	return i.Set(ctx, key, string(data))
}
