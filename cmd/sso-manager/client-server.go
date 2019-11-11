package main

import (
	"context"

	"github.com/gamaops/mono-sso/pkg/datastore"
	ssocommon "github.com/gamaops/mono-sso/pkg/idl/sso-common"
	sso "github.com/gamaops/mono-sso/pkg/idl/sso-manager"
	"github.com/gamaops/mono-sso/pkg/session"
	"github.com/spf13/viper"
)

type ClientServer struct {
	sso.UnimplementedClientServiceServer
}

var upsertClientSession *session.SessionRequestValidator
var upsertScopeSession *session.SessionRequestValidator

func setupClientServer() {
	upsertClientSession = &session.SessionRequestValidator{
		TimestampPastToleration: viper.GetDuration("sessionPastToleration"),
		RequiredScopes: [][]string{
			[]string{"superadmin"},
			[]string{"client:write"},
		},
		AllowedTenants: viper.GetStringSlice("adminTenant"),
		Audience:       viper.GetString("sessionAudience"),
		Jose:           ServiceOAuth2Jose,
	}
	upsertClientSession.Load()

	upsertScopeSession = &session.SessionRequestValidator{
		TimestampPastToleration: viper.GetDuration("sessionPastToleration"),
		RequiredScopes: [][]string{
			[]string{"superadmin"},
			[]string{"scope:write"},
		},
		AllowedTenants: viper.GetStringSlice("adminTenant"),
		Audience:       viper.GetString("sessionAudience"),
		Jose:           ServiceOAuth2Jose,
	}
	upsertScopeSession.Load()
}

func (s *ClientServer) UpsertClient(ctx context.Context, req *sso.UpsertClientRequest) (*sso.UpsertClientResponse, error) {

	res := &sso.UpsertClientResponse{
		Status: &ssocommon.ResponseStatus{},
	}

	if enableSessionValidation {
		_, err := upsertClientSession.ValidateSession(req.Session)
		if err != nil {
			upsertClientSession.ParseSessionErrorToStatus(err, res.Status)
			return res, nil
		}
	}

	client, err := ServiceDatastore.UpsertClient(ctx, req)

	if err != nil {
		log.Errorf("Upsert client error: %v", err)
		datastore.ParseErrorIntoStatus(err, res.Status)
		return res, nil
	}

	res.ClientId = client.ID
	res.ClientSecret = client.Secret

	return res, nil

}

func (s *ClientServer) UpsertScope(ctx context.Context, req *sso.UpsertScopeRequest) (*sso.UpsertScopeResponse, error) {

	res := &sso.UpsertScopeResponse{
		Status: &ssocommon.ResponseStatus{},
	}

	if enableSessionValidation {
		_, err := upsertScopeSession.ValidateSession(req.Session)
		if err != nil {
			upsertScopeSession.ParseSessionErrorToStatus(err, res.Status)
			return res, nil
		}
	}

	_, err := ServiceDatastore.UpsertScope(ctx, req)

	if err != nil {
		log.Errorf("Upsert scope error: %v", err)
		datastore.ParseErrorIntoStatus(err, res.Status)
		return res, nil
	}

	return res, nil

}
