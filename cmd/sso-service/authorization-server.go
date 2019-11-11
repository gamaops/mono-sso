package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"database/sql"

	"github.com/gamaops/mono-sso/pkg/datastore"
	ssomanager "github.com/gamaops/mono-sso/pkg/idl/sso-manager"
	sso "github.com/gamaops/mono-sso/pkg/idl/sso-service"
)

var secretDecoder = base64.StdEncoding.WithPadding(base64.NoPadding)

type AuthorizationServer struct {
	sso.UnimplementedAuthorizationServiceServer
}

func (s *AuthorizationServer) AuthorizeClient(ctx context.Context, req *sso.AuthorizeClientRequest) (*sso.AuthorizeClientResponse, error) {

	res := &sso.AuthorizeClientResponse{}

	isInTenant, err := ServiceDatastore.IsAccountInTenant(ctx, req.Subject, req.TenantId)
	if err != nil {
		res.Status = InternalErrorStatus
		return res, nil
	} else if !isInTenant {
		log.Warnf("Grant request with invalid tenant %v for subject %v", req.TenantId, req.Subject)
		res.Status = InvalidTenantStatus
		return res, nil
	}

	if len(req.Scopes) > 0 {
		hasUkn, err := hasUnknownScopes(ctx, req.ClientId, req.Scopes)
		if err != nil {
			log.Errorf("Error when getting subject's unknown scopes: %v", err)
			res.Status = InternalErrorStatus
			return res, nil
		}

		if hasUkn {
			log.Warnf("Authorization request with unknown scopes: client_id %v scopes %v", req.ClientId, req.Scopes)
			res.Status = UnknownScopesStatus
			return res, nil
		}
	}

	clientName, clientType, err := getAuthorizationClient(ctx, req.ClientId, req.RedirectUri)
	if err != nil {
		log.Errorf("Error when getting authorization to client: %v", err)
		res.Status = InternalErrorStatus
		return res, nil
	}

	if len(clientName) == 0 {
		log.Warnf("Invalid client or redirect uri: %v (%v)", req.ClientId, req.RedirectUri)
		res.Status = InvalidClientStatus
		return res, nil
	}

	if clientType == ssomanager.ClientType_PUBLIC && req.ResponseType == "code" {
		log.Warnf("Invalid response type for this type of client: %v (%v)", req.ClientId, req.RedirectUri)
		res.Status = InvalidResponseTypeStatus
		return res, nil
	}

	res.ClientName = clientName

	unauthScopes, err := getUnauthorizedScopes(ctx, req.ClientId, req.Subject, req.Scopes)
	if err != nil {
		log.Errorf("Error when getting unauthorized scopes: %v", err)
		res.Status = InternalErrorStatus
		return res, nil
	}

	if len(unauthScopes) > 0 {
		res.UnauthorizedScopes = unauthScopes
	}

	return res, nil
}

func (s *AuthorizationServer) GrantScopes(ctx context.Context, req *sso.GrantScopesRequest) (*sso.GrantScopesResponse, error) {

	res := &sso.GrantScopesResponse{}

	if len(req.Scopes) == 0 {
		res.Status = InvalidGrantStatus
		return res, nil
	}

	isInTenant, err := ServiceDatastore.IsAccountInTenant(ctx, req.Subject, req.TenantId)
	if err != nil {
		res.Status = InternalErrorStatus
		return res, nil
	} else if !isInTenant {
		log.Warnf("Grant request with invalid tenant %v for subject %v", req.TenantId, req.Subject)
		res.Status = InvalidTenantStatus
		return res, nil
	}

	scopes, err := getUnauthorizedScopes(ctx, req.ClientId, req.Subject, req.Scopes)
	if err != nil {
		log.Errorf("Error when getting unauthorized scopes: %v", err)
		res.Status = InternalErrorStatus
		return res, nil
	}

	if len(scopes) == 0 || len(scopes) != len(req.Scopes) {
		res.Status = InvalidGrantStatus
		return res, nil
	}

	args := []interface{}{
		req.ClientId,
	}

	for _, scope := range scopes {
		args = append(args, scope)
	}

	result, err := ServiceDatastore.Client.QueryContext(
		ctx,
		`SELECT id FROM sso.scope WHERE client_id = $1 AND deleted_at IS NULL AND scope IN (`+datastore.CreatePlaceholders(1, len(scopes), nil).String()+`)`,
		args...,
	)
	if err != nil {
		log.Errorf("Error when getting scopes IDs: %v", err)
		res.Status = InternalErrorStatus
		return res, nil
	}

	args = []interface{}{
		req.Subject,
	}

	// TODO: Update and remove deleted_at where scopes already granted once

	query := &strings.Builder{}
	query.WriteString("INSERT INTO sso.grant (account_id, scope_id, created_at, updated_at, deleted_at) VALUES ")

	for result.Next() {
		var scopeID string
		if err := result.Scan(&scopeID); err != nil {
			log.Errorf("Error when scanning scope ID: %v", err)
			res.Status = InternalErrorStatus
			return res, nil
		}
		args = append(args, scopeID)
		if len(args) > 2 {
			query.WriteRune(',')
		}
		query.WriteString("($1, $")
		query.WriteString(strconv.FormatInt(int64(len(args)), 10))
		query.WriteString(", now(), now(), null)")
	}

	query.WriteString(" ON CONFLICT ON CONSTRAINT unq_sso_grant_scope_id_account_id DO UPDATE SET deleted_at = null, updated_at = now()")

	_, err = ServiceDatastore.Client.ExecContext(
		ctx,
		query.String(),
		args...,
	)
	if err != nil {
		log.Errorf("Error while granting scopes: %v", err)
		res.Status = InternalErrorStatus
		return res, nil
	}

	ServiceDatastore.RegisterEvent(&sso.RegisterEventRequest{
		Level:       sso.EventLevel_INFO,
		IsSensitive: true,
		Message:     fmt.Sprintf("client requested grants (subject %v): %v", req.Subject, req.ClientId),
		Data: map[string]string{
			"account_id": req.ClientId,
			"client_id":  req.ClientId,
			"scopes":     strings.Join(req.Scopes, ", "),
		},
	})

	return res, nil
}

func (s *AuthorizationServer) NewRefreshToken(ctx context.Context, req *sso.NewRefreshTokenRequest) (*sso.NewRefreshTokenResponse, error) {

	res := &sso.NewRefreshTokenResponse{}

	result := ServiceDatastore.Client.QueryRowContext(
		ctx,
		`SELECT secret FROM sso.client WHERE deleted_at IS NULL AND id = $1`,
		req.ClientId,
	)

	var clientSecret string
	err := result.Scan(&clientSecret)

	if err != nil {
		log.Warnf("Invalid client ID to generate refresh token: %v", err)
		res.Status = InvalidClientStatus
		return res, nil
	}

	reqSecretBytes := []byte(req.ClientSecret)
	clientSecretRaw := make([]byte, secretDecoder.DecodedLen(len(reqSecretBytes)))
	decodeLen, err := secretDecoder.Decode(clientSecretRaw, reqSecretBytes)
	if err != nil {
		switch terr := err.(type) {
		case base64.CorruptInputError:
			log.Warnf("Invalid client secret to generate refresh token: %v", terr)
			res.Status = InvalidClientStatus
			return res, nil
		default:
			log.Warnf("Error when decoding client secret: %v", err)
			res.Status = InvalidClientStatus
			return res, nil
		}
	}

	err = bcrypt.CompareHashAndPassword([]byte(clientSecret), clientSecretRaw[:decodeLen])

	if err != nil {
		log.Warnf("Invalid client secret to generate refresh token: %v", err)
		res.Status = InvalidClientStatus
		return res, nil
	}

	unauthScopes, err := getUnauthorizedScopes(ctx, req.ClientId, req.Subject, req.Scopes)
	if err != nil {
		log.Errorf("Error when getting unauthorized scopes: %v", err)
		res.Status = InternalErrorStatus
		return res, nil
	}

	if len(unauthScopes) > 0 {
		res.Status = UnauthorizedScopesStatus
		return res, nil
	}

	if !req.ForceNew {
		result = ServiceDatastore.Client.QueryRowContext(
			ctx,
			`SELECT id, expires_at FROM sso.token WHERE account_id = $1 AND client_id = $2 AND type = 0 AND deleted_at IS NULL AND expires_at > now()`,
			req.Subject,
			req.ClientId,
		)

		var tokenID string
		var expiresAt time.Time

		err := result.Scan(&tokenID, &expiresAt)
		if err != sql.ErrNoRows {
			if err != nil {
				log.Errorf("Error while getting current refresh token: %v", err)
				res.Status = InternalErrorStatus
				return res, nil
			}

			res.RefreshTokenId = tokenID
			res.ExpiresAt = expiresAt.Unix()
			return res, nil
		}

		if len(req.AuthorizationCode) == 0 {
			res.Status = NoValidRefreshTokenStatus
			return res, nil
		}
	}

	tx, err := ServiceDatastore.Client.BeginTx(ctx, nil)
	if err != nil {
		log.Errorf("Error when starting transaction: %v", err)
		res.Status = InternalErrorStatus
		return res, nil
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE sso.token SET deleted_at = now() WHERE account_id = $1 AND client_id = $2 AND type = 0", req.Subject, req.ClientId)
	if err != nil {
		log.Errorf("Error when invalidating previous refresh tokens: %v", err)
		res.Status = InternalErrorStatus
		return res, nil
	}

	tokenID, err := tokenIDGenerator.New()
	if err != nil {
		log.Errorf("Error when generating refresh token ID: %v", err)
		res.Status = InternalErrorStatus
		return res, nil
	}

	tokenIDStr := tokenID.Base32()
	tokenDuration, _ := time.ParseDuration(req.Duration)
	expiresAt := time.Now().Add(tokenDuration)

	tx.Exec(
		`INSERT INTO sso.token (id, account_id, client_id, type, expires_at, created_at, updated_at, deleted_at)
		VALUES ($1, $2, $3, 0, $4, now(), now(), null)`,
		tokenIDStr,
		req.Subject,
		req.ClientId,
		expiresAt,
	)

	err = tx.Commit()
	if err != nil {
		log.Errorf("Error when committing transaction to create new refresh token: %v", err)
		res.Status = InternalErrorStatus
		return res, nil
	}

	res.ExpiresAt = expiresAt.Unix()

	ServiceDatastore.RegisterEvent(&sso.RegisterEventRequest{
		Level:       sso.EventLevel_INFO,
		IsSensitive: true,
		Message:     fmt.Sprintf("generated new refresh token for account %v and client %v", req.Subject, req.ClientId),
		Data: map[string]string{
			"type":               "0",
			"client_id":          req.ClientId,
			"subject":            req.Subject,
			"session_id":         req.SessionId,
			"redirect_uri":       req.RedirectUri,
			"authorization_code": req.AuthorizationCode,
			"expires_at":         expiresAt.Format(time.RFC3339),
		},
	})

	res.RefreshTokenId = tokenIDStr

	return res, nil

}
