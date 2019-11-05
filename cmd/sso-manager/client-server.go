package main

import (
	"context"

	"github.com/gamaops/mono-sso/pkg/datastore"
	ssocommon "github.com/gamaops/mono-sso/pkg/idl/sso-common"
	sso "github.com/gamaops/mono-sso/pkg/idl/sso-manager"
)

type ClientServer struct {
	sso.UnimplementedClientServiceServer
}

func (s *ClientServer) UpsertClient(ctx context.Context, req *sso.UpsertClientRequest) (*sso.UpsertClientResponse, error) {

	res := &sso.UpsertClientResponse{
		Status: &ssocommon.ResponseStatus{},
	}

	client, err := ServiceDatastore.UpsertClient(ctx, req)

	if err != nil {
		log.Errorf("Upser client error: %v", err)
		datastore.ParseErrorIntoStatus(err, res.Status)
		return res, nil
	}

	res.ClientId = client.ID.Hex()

	return res, nil

}
