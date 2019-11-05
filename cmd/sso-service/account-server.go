package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/gamaops/mono-sso/pkg/datastore"
	sso "github.com/gamaops/mono-sso/pkg/idl/sso-service"
	"golang.org/x/crypto/bcrypt"
)

type AccountServer struct {
	sso.UnimplementedAccountServiceServer
}

func (s *AccountServer) SignIn(ctx context.Context, req *sso.SignInRequest) (*sso.SignInResponse, error) {
	result := ServiceDatastore.Client.QueryRowContext(
		ctx,
		`SELECT acc.id, acc.name, acc.activation_method, acc.password
		FROM sso.account AS acc
		INNER JOIN sso.account_identifier AS acc_id ON (acc_id.account_id = acc.id AND acc_id.identifier = $1)
		WHERE acc.deleted_at IS NULL`,
		req.Identifier,
	)

	acc := &datastore.AccountDoc{}

	err := result.Scan(
		&acc.ID,
		&acc.Name,
		&acc.ActivationMethod,
		&acc.Password,
	)

	res := &sso.SignInResponse{}
	if err != nil {
		log.Debugf("Invalid account identifier/password: %v", err)
		res.Status = SignInInvalidAccountStatus
		return res, nil
	}

	err = bcrypt.CompareHashAndPassword([]byte(acc.Password), []byte(req.Password))

	if err != nil {
		log.Debugf("Invalid account identifier/password: %v", err)
		res.Status = SignInInvalidAccountStatus
		return res, nil
	}

	res.ActivationMethod = acc.ActivationMethod
	res.Subject = acc.ID
	res.Profile = &sso.AccountProfile{
		Name: acc.Name,
	}

	ServiceDatastore.RegisterEvent(&sso.RegisterEventRequest{
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

	ServiceDatastore.RegisterEvent(req)

	return res, nil
}
