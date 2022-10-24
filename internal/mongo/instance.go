package mongo

import (
	"context"
	"net/url"

	"github.com/JoachimFlottorp/magnolia/internal/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CollectionName string

const (
	CollectionAPILog = CollectionName("api_log")
	CollectionTwitch = CollectionName("twitch")
)

var ErrNoDocuments = mongo.ErrNoDocuments

type Instance interface {
	Collection(CollectionName) *mongo.Collection
	Ping(ctx context.Context) error
	RawClient() *mongo.Client
	RawDatabase() *mongo.Database
}

type mongoInst struct {
	client *mongo.Client
	db     *mongo.Database
}

func createUrl(cfg *config.Config) string {
	uri := "mongodb"
	if cfg.Mongo.SRV {
		uri += "+srv"
	}
	uri += "://"
	if cfg.Mongo.Username != "" {
		uri += url.QueryEscape(cfg.Mongo.Username)
		if cfg.Mongo.Password != "" {
			uri += ":" + url.QueryEscape(cfg.Mongo.Password)
		}
		uri += "@"
	}
	uri += cfg.Mongo.Address
	if cfg.Mongo.DB != "" {
		uri += "/" + cfg.Mongo.DB
	}
	return uri
}

func New(ctx context.Context, cfg *config.Config) (Instance, error) {
	uri := createUrl(cfg)
	
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	db := client.Database(cfg.Mongo.DB)

	_ = db.CreateCollection(ctx, string(CollectionAPILog))
	_ = db.CreateCollection(ctx, string(CollectionTwitch))

	return &mongoInst{
		client: client,
		db:     db,
	}, nil
}

func (i *mongoInst) Collection(name CollectionName) *mongo.Collection {
	return i.db.Collection(string(name))
}

func (i *mongoInst) Ping(ctx context.Context) error {
	return i.db.Client().Ping(ctx, nil)
}

func (i *mongoInst) RawClient() *mongo.Client {
	return i.client
}

func (i *mongoInst) RawDatabase() *mongo.Database {
	return i.db
}
