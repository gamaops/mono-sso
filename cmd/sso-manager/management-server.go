package main

import (
	"context"

	"github.com/gamaops/mono-sso/pkg/datastore"
	ssocommon "github.com/gamaops/mono-sso/pkg/idl/sso-common"
	sso "github.com/gamaops/mono-sso/pkg/idl/sso-manager"
	"github.com/gamaops/mono-sso/pkg/session"
	"github.com/spf13/viper"
)

type ManagementServer struct {
	sso.UnimplementedManagementServiceServer
}

var upsertTenantSession *session.SessionRequestValidator

func setupManagementServer() {
	upsertTenantSession = &session.SessionRequestValidator{
		TimestampPastToleration: viper.GetDuration("sessionPastToleration"),
		RequiredScopes: [][]string{
			[]string{"superadmin"},
			[]string{"tenant:write"},
		},
		AllowedTenants: viper.GetStringSlice("adminTenant"),
		Audience:       viper.GetString("sessionAudience"),
		Jose:           ServiceOAuth2Jose,
	}
	upsertTenantSession.Load()
}

func (s *ManagementServer) UpsertTenant(ctx context.Context, req *sso.UpsertTenantRequest) (*sso.UpsertTenantResponse, error) {

	res := &sso.UpsertTenantResponse{
		Status: &ssocommon.ResponseStatus{},
	}

	if enableSessionValidation {
		_, err := upsertTenantSession.ValidateSession(req.Session)
		if err != nil {
			upsertTenantSession.ParseSessionErrorToStatus(err, res.Status)
			return res, nil
		}
	}

	tenant, err := ServiceDatastore.UpsertTenant(ctx, req)

	if err != nil {
		log.Errorf("Upsert tenant error: %v", err)
		datastore.ParseErrorIntoStatus(err, res.Status)
		return res, nil
	}

	res.TenantId = tenant.ID

	return res, nil

}
