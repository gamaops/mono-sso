package datastore

import (
	"time"

	sso "github.com/gamaops/mono-sso/pkg/idl/sso-service"
)

type AccountDoc struct {
	ID               string
	Password         string
	ActivationMethod sso.ActivationMethod
	Name             string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        time.Time
}
