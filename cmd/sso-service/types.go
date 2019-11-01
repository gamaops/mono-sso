package main

import (
	"time"

	ssomanager "github.com/gamaops/mono-sso/pkg/idl/sso-manager"
	sso "github.com/gamaops/mono-sso/pkg/idl/sso-service"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AccountEntity struct {
	ID               *primitive.ObjectID  `bson:"_id,omitempty"`
	Identifier       string               `bson:"identifier"`
	Password         string               `bson:"password"`
	ActivationMethod sso.ActivationMethod `bson:"activation_method"`
	Enabled          bool                 `bson:"enabled"`
	Name             string               `bson:"name"`
}

type ClientEntity struct {
	ID           *primitive.ObjectID   `bson:"_id,omitempty"`
	Name         string                `bson:"name"`
	Secret       string                `bson:"secret"`
	Type         ssomanager.ClientType `bson:"type"`
	RedirectURIs []string              `bson:"redirect_uris"`
}

type GrantScopesDifference struct {
	UnauthorizedScopes []string `bson:"unauthorized_scopes"`
}

type UnknownScopesDifference struct {
	UnknownScopes []string `bson:"unknown_scopes"`
}

type TokenType uint16

const (
	RefreshToken TokenType = 0
)

type RefreshTokenEntity struct {
	ID        *primitive.ObjectID `bson:"_id,omitempty"`
	ExpiresAt time.Time           `bson:"expires_at"`
}
