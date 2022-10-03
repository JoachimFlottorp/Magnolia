package mongo

import (
	"context"
	"fmt"
	"net/url"

	"github.com/JoachimFlottorp/yeahapi/internal/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CollectionName string

const (
	CollectionAPILog CollectionName = "api_log"
)

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

func New(ctx context.Context, cfg *config.Config) (Instance, error) {
	uri := fmt.Sprintf(
		"mongodb+srv://%s:%s@%s", 
		url.QueryEscape(cfg.Mongo.Username), 
		url.QueryEscape(cfg.Mongo.Password), 
		cfg.Mongo.Address,
	)
	
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	db := client.Database(cfg.Mongo.DB)

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
