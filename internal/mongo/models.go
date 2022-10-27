package mongo

import (
	"context"
	"errors"
	"time"

	"github.com/JoachimFlottorp/magnolia/external"
	"github.com/JoachimFlottorp/magnolia/external/ivr"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrMissingName  = errors.New("missing name")
	ErrNameNotExist = errors.New("name does not exist")
)

type ApiLog struct {
	// ID is the unique identifier for the log entry.
	ID primitive.ObjectID `json:"id" bson:"_id"`

	// Timestamp is the time the log entry was created.
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`

	// Method is the HTTP method used for the request.
	Method string `json:"method" bson:"method"`

	// Path is the path used for the request.
	Path string `json:"path" bson:"path"`

	// Status is the HTTP status code returned for the request.
	Status int `json:"status" bson:"status"`

	// IP is the IP address of the client.
	IP string `json:"ip" bson:"ip"`

	// UserAgent is the user agent of the client.
	UserAgent string `json:"user_agent" bson:"user_agent"`

	// Error is the error returned for the request.
	Error string `json:"error,omitempty" bson:"error"`
}

type TwitchChannel struct {
	ID         primitive.ObjectID `json:"id" bson:"_id"`
	TwitchID   string             `json:"twitch_id" bson:"twitch_id"`
	TwitchName string             `json:"twitch_name" bson:"twitch_name"`
}

func (t *TwitchChannel) GetByName(ctx context.Context, i Instance) error {
	return i.Collection(CollectionTwitch).FindOne(ctx, bson.M{"twitch_name": t.TwitchName}).Decode(t)
}

func (t *TwitchChannel) ResolveByIVR(ctx context.Context) error {
	if t.TwitchName == "" {
		return ErrMissingName
	}

	users, err := ivr.ResolveUsernames(ctx, external.Client(), []string{t.TwitchName})
	if err != nil {
		return err
	}

	if len(users) == 0 {
		return ErrNameNotExist
	}

	t.TwitchName = users[0].Login
	t.TwitchID = users[0].ID

	return nil
}

func (t *TwitchChannel) Save(ctx context.Context, i Instance) error {
	if t.ID.IsZero() {
		t.ID = primitive.NewObjectID()
	}

	p := true

	_, err := i.Collection(CollectionTwitch).ReplaceOne(ctx, bson.M{"_id": t.ID}, t, &options.ReplaceOptions{
		Upsert: &p,
	})

	return err
}
