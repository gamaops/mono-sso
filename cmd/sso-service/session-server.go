package main

import (
	"context"
	"fmt"
	"strconv"

	sso "github.com/gamaops/mono-sso/pkg/idl/sso-service"
)

type SessionServer struct {
	sso.UnimplementedSessionServiceServer
}

func (s *SessionServer) PurgeClientCache(ctx context.Context, req *sso.PurgeClientCacheRequest) (*sso.PurgeClientCacheResponse, error) {

	res := &sso.PurgeClientCacheResponse{}

	count, err := ServiceCache.PurgeClientCache(req.ClientId)

	if err != nil {
		log.Errorf("Error when cleaning client cache: %v", err)
		res.Status = InternalErrorStatus
		return res, nil
	}

	res.DeletedCount = count

	ServiceDatastore.RegisterEvent(&sso.RegisterEventRequest{
		Level:       sso.EventLevel_WARNING,
		IsSensitive: false,
		Message:     fmt.Sprintf("client cache cleaned: %v", req.ClientId),
		Data: map[string]string{
			"client_id": req.ClientId,
			"count":     strconv.FormatInt(int64(count), 10),
		},
	})

	return res, nil
}

func (s *SessionServer) PurgeAccountCache(ctx context.Context, req *sso.PurgeAccountCacheRequest) (*sso.PurgeAccountCacheResponse, error) {

	res := &sso.PurgeAccountCacheResponse{}

	count, err := ServiceCache.PurgeSubjectCache(req.Subject, req.SessionId)

	if err != nil {
		log.Errorf("Error when cleaning subject cache: %v", err)
		res.Status = InternalErrorStatus
		return res, nil
	}

	res.DeletedCount = count

	ServiceDatastore.RegisterEvent(&sso.RegisterEventRequest{
		Level:       sso.EventLevel_WARNING,
		IsSensitive: true,
		Message:     fmt.Sprintf("subject cache cleaned: %v", req.Subject),
		Data: map[string]string{
			"account_id": req.Subject,
			"session":    req.SessionId,
			"count":      strconv.FormatInt(int64(count), 10),
		},
	})

	return res, nil
}
