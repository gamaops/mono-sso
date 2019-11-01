package main

import (
	"context"
	"io"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func getUnauthorizedScopes(ctx context.Context, clientID primitive.ObjectID, subject primitive.ObjectID, scopes []string) (*GrantScopesDifference, error) {
	query := bson.A{
		bson.M{
			"$match": bson.M{
				"client_id": bson.M{
					"$eq": clientID,
				},
				"subject": bson.M{
					"$eq": subject,
				},
			},
		},
		bson.M{
			"$project": bson.M{
				"unauthorized_scopes": bson.M{
					"$setDifference": bson.A{
						scopes,
						"$scopes",
					},
				},
			},
		},
	}
	cursor, err := ServiceDatastore.Collections.Grants.Aggregate(ctx, query)

	if err != nil {
		log.Errorf("Error while getting scopes difference on grants: %v", err)
		return nil, err
	}

	diff := &GrantScopesDifference{}
	more := cursor.Next(ctx)
	err = cursor.Decode(diff)
	cursor.Close(ctx)

	if !more {
		diff.UnauthorizedScopes = scopes
	} else if err != nil && err == io.EOF {
		log.Errorf("Error while decoding scopes difference on grants: %v", err)
		return nil, err
	}

	return diff, nil

}

func getUnknownScopes(ctx context.Context, clientID primitive.ObjectID, scopes []string) (*UnknownScopesDifference, error) {
	query := bson.A{
		bson.M{
			"$match": bson.M{
				"client_id": bson.M{
					"$eq": clientID,
				},
				"scope": bson.M{
					"$in": scopes,
				},
			},
		},
		bson.M{
			"$group": bson.M{
				"_id": nil,
				"found_scopes": bson.M{
					"$addToSet": "$scope",
				},
			},
		},
		bson.M{
			"$project": bson.M{
				"unknown_scopes": bson.M{
					"$setDifference": bson.A{
						scopes,
						"$found_scopes",
					},
				},
			},
		},
	}
	cursor, err := ServiceDatastore.Collections.Scopes.Aggregate(ctx, query)

	if err != nil {
		log.Errorf("Error while getting unknown scopes: %v", err)
		return nil, err
	}

	diff := &UnknownScopesDifference{}
	more := cursor.Next(ctx)
	err = cursor.Decode(diff)
	cursor.Close(ctx)

	if !more {
		diff.UnknownScopes = scopes
	} else if err != nil && err != io.EOF {
		log.Errorf("Error while decoding unknown scopes: %v", err)
		return nil, err
	}

	return diff, nil

}

func getAuthorizationClient(ctx context.Context, clientID primitive.ObjectID, redirectUri string) (*ClientEntity, error) {
	query := bson.M{
		"_id": bson.M{
			"$eq": clientID,
		},
		"redirect_uris": bson.M{
			"$eq": redirectUri,
		},
	}

	result := ServiceDatastore.Collections.Clients.FindOne(ctx, query, ClientAuthorizationOpts)

	err := result.Err()

	// TODO: Handle "mongo: no documents in result"
	if err != nil || err == mongo.ErrNoDocuments {
		log.Warnf("Error while getting authorization client from MongoDB: %v", err)
		return nil, err
	}

	client := &ClientEntity{}

	err = result.Decode(client)

	if err != nil {
		log.Errorf("Error while decoding authorization client from MongoDB: %v", err)
		return nil, err
	}

	return client, nil

}
