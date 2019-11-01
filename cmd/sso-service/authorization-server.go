package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	ssomanager "github.com/gamaops/mono-sso/pkg/idl/sso-manager"
	sso "github.com/gamaops/mono-sso/pkg/idl/sso-service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthorizationServer struct {
	sso.UnimplementedAccountServiceServer
}

func (s *AuthorizationServer) AuthorizeClient(ctx context.Context, req *sso.AuthorizeClientRequest) (*sso.AuthorizeClientResponse, error) {

	res := &sso.AuthorizeClientResponse{}

	clientID, err := primitive.ObjectIDFromHex(req.ClientId)
	if err != nil {
		log.Errorf("Error when converting client ID to object ID: %v", err)
		res.Status = InternalErrorStatus
		return res, nil
	}
	subject, _ := primitive.ObjectIDFromHex(req.Subject)
	if err != nil {
		log.Errorf("Error when converting client ID to object ID: %v", err)
		res.Status = InternalErrorStatus
		return res, nil
	}

	if len(req.Scopes) > 0 {
		diffUkn, err := getUnknownScopes(ctx, clientID, req.Scopes)
		if err != nil {
			log.Errorf("Error when getting subject's unknown scopes: %v", err)
			res.Status = InternalErrorStatus
			return res, nil
		}

		if len(diffUkn.UnknownScopes) > 0 {
			log.Warnf("Authorization request with unknown scopes: %v", diffUkn.UnknownScopes)
			res.Status = UnknownScopesStatus
			return res, nil
		}
	}

	client, err := getAuthorizationClient(ctx, clientID, req.RedirectUri)
	if err != nil {
		log.Errorf("Error when getting authorization to client: %v", err)
		res.Status = InternalErrorStatus
		return res, nil
	}

	if client == nil || len(client.Name) == 0 {
		log.Warnf("Invalid client or redirect uri: %v (%v)", req.ClientId, req.RedirectUri)
		res.Status = InvalidClientStatus
		return res, nil
	}

	if client.Type == ssomanager.ClientType_PUBLIC && req.ResponseType == "code" {
		log.Warnf("Invalid response type for this type of client: %v (%v)", req.ClientId, req.RedirectUri)
		res.Status = InvalidResponseTypeStatus
		return res, nil
	}

	res.ClientName = client.Name

	diffUna, err := getUnauthorizedScopes(ctx, clientID, subject, req.Scopes)
	if err != nil {
		log.Errorf("Error when getting unauthorized scopes: %v", err)
		res.Status = InternalErrorStatus
		return res, nil
	}

	if len(diffUna.UnauthorizedScopes) > 0 {
		res.UnauthorizedScopes = diffUna.UnauthorizedScopes
	}

	return res, nil
}

func (s *AuthorizationServer) GrantScopes(ctx context.Context, req *sso.GrantScopesRequest) (*sso.GrantScopesResponse, error) {

	res := &sso.GrantScopesResponse{}

	clientID, err := primitive.ObjectIDFromHex(req.ClientId)
	if err != nil {
		log.Errorf("Error when converting client ID to object ID: %v", err)
		res.Status = InternalErrorStatus
		return res, nil
	}
	subject, err := primitive.ObjectIDFromHex(req.Subject)
	if err != nil {
		log.Errorf("Error when converting client ID to object ID: %v", err)
		res.Status = InternalErrorStatus
		return res, nil
	}

	_, err = ServiceDatastore.Collections.Grants.UpdateOne(ctx, bson.M{
		"subject": bson.M{
			"$eq": subject,
		},
		"client_id": bson.M{
			"$eq": clientID,
		},
	}, bson.M{
		"$addToSet": bson.M{
			"scopes": bson.M{
				"$each": req.Scopes,
			},
		},
		"$set": bson.M{
			"subject":   subject,
			"client_id": clientID,
		},
	}, GrantUpdateOptions)

	if err != nil {
		log.Errorf("Error while granting scopes: %v", err)
		res.Status = InternalErrorStatus
		return res, nil
	}

	err = ServiceDatastore.InsertEvent(ctx, &sso.RegisterEventRequest{
		Level:       sso.EventLevel_INFO,
		IsSensitive: true,
		Message:     fmt.Sprintf("client requested grants (subject %v): %v", req.Subject, req.ClientId),
		Data: map[string]string{
			"account_id": req.ClientId,
			"client_id":  req.ClientId,
			"scopes":     strings.Join(req.Scopes, ", "),
		},
	})

	if err != nil {
		log.Errorf("Error when inserting event: %v", err)
		res.Status = InternalErrorStatus
		return res, nil
	}

	return res, nil
}

func (s *AuthorizationServer) NewRefreshToken(ctx context.Context, req *sso.NewRefreshTokenRequest) (*sso.NewRefreshTokenResponse, error) {

	res := &sso.NewRefreshTokenResponse{}

	clientID, _ := primitive.ObjectIDFromHex(req.ClientId)

	clientResult := ServiceDatastore.Collections.Clients.FindOne(ctx, bson.M{
		"_id": bson.M{
			"$eq": clientID,
		},
	}, FindClientRefreshTokenOpts)

	err := clientResult.Err()

	if err != nil {
		log.Warnf("Invalid client ID to generate refresh token: %v", err)
		res.Status = InvalidClientStatus
		return res, nil
	}

	client := &ClientEntity{}
	err = clientResult.Decode(client)
	if err != nil {
		log.Errorf("Error while decoding client for refresh token: %v", err)
		res.Status = InternalErrorStatus
		return res, nil
	}

	err = bcrypt.CompareHashAndPassword([]byte(client.Secret), []byte(req.ClientSecret))

	if err != nil {
		log.Warnf("Invalid client secret to generate refresh token: %v", err)
		res.Status = InvalidClientStatus
		return res, nil
	}

	subject, _ := primitive.ObjectIDFromHex(req.Subject)

	findQuery := bson.M{
		"type": bson.M{
			"$eq": RefreshToken,
		},
		"client_id": bson.M{
			"$eq": clientID,
		},
		"subject": bson.M{
			"$eq": subject,
		},
		"enabled": bson.M{
			"$eq": true,
		},
		"expires_at": bson.M{
			"$gt": time.Now(),
		},
	}

	if !req.ForceNew {
		refreshTokenResult := ServiceDatastore.Collections.Tokens.FindOne(ctx, findQuery, FindRefreshTokenOpts)

		err := refreshTokenResult.Err()
		if err != mongo.ErrNoDocuments {
			if err != nil {
				log.Errorf("Error while getting current refresh token: %v", err)
				res.Status = InternalErrorStatus
				return res, nil
			}

			refreshToken := &RefreshTokenEntity{}
			err = refreshTokenResult.Decode(refreshToken)
			if err != nil {
				log.Errorf("Error while decoding current refresh token: %v", err)
				res.Status = InternalErrorStatus
				return res, nil
			}

			res.RefreshTokenId = refreshToken.ID.Hex()
			res.ExpiresAt = refreshToken.ExpiresAt.Unix()
			return res, nil
		}

		if len(req.AuthorizationCode) == 0 {
			res.Status = NoValidRefreshTokenStatus
			return res, nil
		}
	}

	delete(findQuery, "expires_at")

	ServiceDatastore.Collections.Tokens.UpdateMany(ctx, findQuery, bson.M{
		"$set": bson.M{
			"enabled": false,
		},
	})

	tokenDuration, _ := time.ParseDuration(req.Duration)
	expiresAt := time.Now().Add(tokenDuration)
	sessionID, _ := primitive.ObjectIDFromHex(req.SessionId)

	res.ExpiresAt = expiresAt.Unix()

	newRefreshToken, err := ServiceDatastore.Collections.Tokens.InsertOne(ctx, bson.M{
		"type":               RefreshToken,
		"client_id":          clientID,
		"subject":            subject,
		"session_id":         sessionID,
		"redirect_uri":       req.RedirectUri,
		"authorization_code": req.AuthorizationCode,
		"enabled":            true,
		"created_at":         time.Now(),
		"expires_at":         expiresAt,
	})

	if err != nil {
		log.Errorf("Error while inserting new refresh token: %v", err)
		res.Status = InternalErrorStatus
		return res, nil
	}

	refreshTokenID := newRefreshToken.InsertedID.(primitive.ObjectID)

	res.RefreshTokenId = refreshTokenID.Hex()

	return res, nil

}
