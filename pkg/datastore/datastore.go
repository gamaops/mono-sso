package datastore

import (
	"context"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Options struct {
	MongoDBURI    string
	MongoDatabase string
}

type MongoCollections struct {
	Accounts *mongo.Collection
	Events   *mongo.Collection
	Grants   *mongo.Collection
	Scopes   *mongo.Collection
	Tokens   *mongo.Collection
	Clients  *mongo.Collection
}

type Datastore struct {
	Options     *Options
	Client      *mongo.Client
	Database    *mongo.Database
	Collections *MongoCollections
	Logger      *logrus.Logger
}

func StartDatastore(ctx context.Context, datastore *Datastore) error {
	datastore.Logger.Info("Connecting to MongoDB")
	datastore.Collections = &MongoCollections{}
	client, err := mongo.NewClient(options.Client().ApplyURI(datastore.Options.MongoDBURI))
	if err != nil {
		return err
	}

	datastore.Client = client
	err = client.Connect(ctx)
	if err != nil {
		return err
	}

	database := client.Database(datastore.Options.MongoDatabase)

	datastore.Database = database

	datastore.Collections.Accounts = database.Collection("accounts")
	datastore.Collections.Events = database.Collection("events")
	datastore.Collections.Scopes = database.Collection("scopes")
	datastore.Collections.Grants = database.Collection("grants")
	datastore.Collections.Tokens = database.Collection("tokens")
	datastore.Collections.Clients = database.Collection("clients")

	return ensureIndexes(ctx, datastore)
}

func ensureIndexes(ctx context.Context, datastore *Datastore) error {
	datastore.Logger.Info("Ensuring MongoDB indexes")
	var err error = nil

	optsID := options.Index().SetName("unq_identifier").SetUnique(true)

	idxs := []mongo.IndexModel{
		mongo.IndexModel{
			Keys:    bson.M{"identifier": 1},
			Options: optsID,
		},
	}

	_, err = datastore.Collections.Accounts.Indexes().CreateMany(ctx, idxs)
	if err != nil {
		return err
	}

	optsID = options.Index().SetName("unq_scopes").SetUnique(true)

	idxs = []mongo.IndexModel{
		mongo.IndexModel{
			Keys:    bson.M{"scope": 1, "client_id": 1},
			Options: optsID,
		},
	}

	_, err = datastore.Collections.Scopes.Indexes().CreateMany(ctx, idxs)
	if err != nil {
		return err
	}

	optsID = options.Index().SetName("unq_grants").SetUnique(true)

	idxs = []mongo.IndexModel{
		mongo.IndexModel{
			Keys:    bson.M{"client_id": 1, "subject": 1},
			Options: optsID,
		},
	}

	_, err = datastore.Collections.Grants.Indexes().CreateMany(ctx, idxs)
	if err != nil {
		return err
	}

	optsID = options.Index().SetName("unq_refresh_tokens").SetUnique(true).SetPartialFilterExpression(bson.M{
		"type": bson.M{
			"$eq": 0,
		},
		"enabled": bson.M{
			"$eq": true,
		},
	})

	idxs = []mongo.IndexModel{
		mongo.IndexModel{
			Keys:    bson.M{"client_id": 1, "subject": 1},
			Options: optsID,
		},
	}

	_, err = datastore.Collections.Tokens.Indexes().CreateMany(ctx, idxs)
	if err != nil {
		return err
	}

	return nil
}

func StopDataStore(ctx context.Context, datastore *Datastore) {
	datastore.Logger.Warn("Disconnecting from MongoDB")
	if err := datastore.Client.Disconnect(ctx); err != nil {
		datastore.Logger.Errorf("Error while disconnecting from MongoDB: %v", err)
	}
}
