package main

import (
	"context"
	"fmt"
	"strings"

	sso "github.com/gamaops/mono-sso/pkg/idl/sso-service"
	"github.com/rs/xid"
	logrus "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AccountServer struct {
	sso.UnimplementedAccountServiceServer
}

func (s *AccountServer) SignIn(ctx context.Context, req *sso.SignInRequest) (*sso.SignInResponse, error) {
	filter := bson.M{
		"identifier": req.Identifier,
		"enabled":    true,
	}
	result := ServiceDatastore.Collections.Accounts.FindOne(ctx, filter, SignInFindOpts)
	err := result.Err()

	res := &sso.SignInResponse{}
	if err != nil {
		log.Debugf("Invalid account identifier/password: %v", err)
		res.Status = SignInInvalidAccountStatus
		return res, nil
	}

	acc := &AccountEntity{}
	err = result.Decode(acc)
	if err != nil {
		errid := xid.New().String()
		log.WithFields(logrus.Fields{
			"errid": errid,
		}).Errorf("Error while decoding entity from MongoDB: %v", err)
		return nil, status.Errorf(codes.Internal, "Internal error: %v", errid)
	}

	err = bcrypt.CompareHashAndPassword([]byte(acc.Password), []byte(req.Password))

	if err != nil {
		log.Debugf("Invalid account identifier/password: %v", err)
		res.Status = SignInInvalidAccountStatus
		return res, nil
	}

	res.ActivationMethod = acc.ActivationMethod
	res.Subject = acc.ID.Hex()
	res.Profile = &sso.AccountProfile{
		Name: acc.Name,
	}

	ServiceDatastore.InsertEvent(ctx, &sso.RegisterEventRequest{
		Level:       sso.EventLevel_INFO,
		IsSensitive: true,
		Message:     fmt.Sprintf("account signed in (%v): %v", res.Subject, req.Identifier),
		Data: map[string]string{
			"identifier":    req.Identifier,
			"account_id":    res.Subject,
			"user_agent":    req.UserAgent,
			"source_ip":     req.SourceIp,
			"client_ip":     req.ClientIp,
			"forwarded_ips": strings.Join(req.ForwardedIps, ", "),
			"session_id":    req.SessionId,
		},
	})

	err = generateActivationCode(ctx, acc, req)
	if err != nil {
		log.Errorf("Error when generating activation code: %v", err)
		res.Status = InternalErrorStatus
	}

	return res, nil
}

func (s *AccountServer) ActivateSession(ctx context.Context, req *sso.ActivateSessionRequest) (*sso.ActivateSessionResponse, error) {

	res := &sso.ActivateSessionResponse{}

	valid, err := validateActivationCode(ctx, req)
	if err != nil {
		log.Errorf("Error when validating activation code: %v", err)
		res.Status = InternalErrorStatus
	}

	if !valid {
		res.Status = InvalidActivationRequestStatus
	}

	return res, nil
}

func (s *AccountServer) RegisterEvent(ctx context.Context, req *sso.RegisterEventRequest) (*sso.RegisterEventResponse, error) {

	res := &sso.RegisterEventResponse{}

	err := ServiceDatastore.InsertEvent(ctx, req)

	if err != nil {
		log.Errorf("Error when inserting event: %v", err)
		res.Status = InternalErrorStatus
	}

	return res, nil
}
