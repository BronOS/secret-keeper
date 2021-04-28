package db

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type MongoDB struct {
	opts       *options.ClientOptions
	ctx        context.Context
	client     *mongo.Client
	db         *mongo.Database
	collection *mongo.Collection
}

func NewMongoDB(c *Config) *MongoDB {
	return &MongoDB{
		opts: buildClientOptions(c),
		ctx:  context.Background(),
	}
}

func newContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}

func buildClientOptions(c *Config) *options.ClientOptions {
	uri := fmt.Sprintf("mongodb://%s", c.Addr)
	if c.Port > 0 {
		uri += fmt.Sprintf(":%d", c.Port)
	}

	opts := options.Client().ApplyURI(uri)

	if len(c.User) > 0 {
		opts.SetAuth(options.Credential{
			Username: c.User,
			Password: c.Pass,
		})
	}

	return opts
}

func (m *MongoDB) Connect() error {
	var err error
	collectionName := "secrets"

	if m.client, err = mongo.Connect(m.ctx, m.opts); err != nil {
		return err
	}

	if err = m.client.Ping(m.ctx, nil); err != nil {
		return err
	}

	m.db = m.client.Database("secret-keeper")
	m.collection = m.db.Collection(collectionName)

	if !m.isCollectionExists(collectionName) {
		if err := m.prepareCollection(); err != nil {
			return err
		}
	}

	return nil
}

func (m *MongoDB) prepareCollection() error {
	ctx, cancel := newContext()
	defer cancel()

	models := []mongo.IndexModel{{
		Keys: bson.M{
			"key": 1,
		},
		Options: options.Index().SetUnique(true),
	}, {
		Keys: bson.M{
			"exp_ts": 1,
		},
		Options: options.Index().SetExpireAfterSeconds(0),
	}}

	if _, err := m.collection.Indexes().CreateMany(ctx, models); err != nil {
		return err
	}

	return nil
}

func (m *MongoDB) isCollectionExists(name string) bool {
	ctx, cancel := newContext()
	defer cancel()

	names, err := m.db.ListCollectionNames(ctx, nil)
	if err != nil {
		return false
	}

	for _, n := range names {
		if n == name {
			return true
		}
	}

	return false
}

func (m *MongoDB) Disconnect() error {
	return m.client.Disconnect(m.ctx)
}

func (m *MongoDB) Set(s *SecretSchema) error {
	ctx, cancel := newContext()
	defer cancel()

	if _, err := m.collection.InsertOne(ctx, s); err != nil {
		return err
	}

	return nil
}

func (m *MongoDB) Get(key string) (*SecretSchema, error) {
	ctx, cancel := newContext()
	defer cancel()

	s := &SecretSchema{}
	if err := m.collection.FindOne(ctx, bson.M{
		"key": key,
	}).Decode(s); err != nil {
		return nil, err
	}
	return s, nil
}

func (m *MongoDB) IncNumTries(key string) error {
	ctx, cancel := newContext()
	defer cancel()

	if _, err := m.collection.UpdateOne(ctx, bson.M{
		"key": key,
	}, bson.D{
		{"$inc", bson.D{{"num_tries", 1}}},
	}); err != nil {
		return err
	}

	return nil
}

func (m *MongoDB) Delete(key string) error {
	ctx, cancel := newContext()
	defer cancel()

	res, err := m.collection.DeleteOne(ctx, bson.M{
		"key": key,
	})

	if err != nil {
		return err
	}

	if res.DeletedCount == 0 {
		return errors.New("secret not found")
	}

	return nil
}
