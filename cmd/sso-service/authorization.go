package main

import (
	"context"

	"github.com/gamaops/mono-sso/pkg/datastore"
	ssomanager "github.com/gamaops/mono-sso/pkg/idl/sso-manager"
)

func getUnauthorizedScopes(ctx context.Context, clientID string, subject string, scopes []string) ([]string, error) {

	stmt, err := ServiceDatastore.Client.PrepareContext(
		ctx,
		`SELECT rscope
		FROM unnest(ARRAY[`+datastore.CreatePlaceholders(2, len(scopes), nil).String()+`]) rscope
		INNER JOIN sso.scope AS sco ON (sco.scope = rscope AND sco.client_id = $1 AND sco.deleted_at IS NULL)
		LEFT JOIN sso.grant AS grt ON (grt.scope_id = sco.id AND grt.account_id = $2 AND grt.deleted_at IS NULL)
		WHERE grt.created_at IS NULL`,
	)
	if err != nil {
		log.Errorf("Error while preparing statement: %v", err)
		return nil, err
	}
	defer stmt.Close()

	args := []interface{}{
		clientID,
		subject,
	}
	for _, scope := range scopes {
		args = append(args, scope)
	}

	result, err := stmt.Query(args...)
	if err != nil {
		log.Errorf("Error while executing prepared statement: %v", err)
		return nil, err
	}
	defer result.Close()

	unauthorizedScopes := make([]string, 0)

	for result.Next() {
		var scope string
		if err := result.Scan(&scope); err != nil {
			return nil, err
		}
		unauthorizedScopes = append(unauthorizedScopes, scope)
	}

	return unauthorizedScopes, nil

}

func hasUnknownScopes(ctx context.Context, clientID string, scopes []string) (bool, error) {

	stmt, err := ServiceDatastore.Client.PrepareContext(
		ctx,
		"SELECT COUNT(1) FROM sso.scope WHERE client_id = $1 AND scope IN ("+datastore.CreatePlaceholders(1, len(scopes), nil).String()+") AND deleted_at IS NULL",
	)
	if err != nil {
		return false, err
	}
	defer stmt.Close()
	args := []interface{}{
		clientID,
	}
	for _, scope := range scopes {
		args = append(args, scope)
	}
	result := stmt.QueryRow(args...)

	var count int = 0

	err = result.Scan(&count)

	if err != nil {
		log.Errorf("Error while getting unknown scopes: %v", err)
		return false, err
	}

	if count != len(scopes) {
		return true, err
	}

	return false, nil

}

func getAuthorizationClient(ctx context.Context, clientID string, redirectURI string) (string, ssomanager.ClientType, error) {

	result := ServiceDatastore.Client.QueryRow(
		`SELECT name, type FROM sso.client WHERE id = $1 AND $2 = ANY(redirect_uris) AND deleted_at IS NULL`,
		clientID,
		redirectURI,
	)

	var clientName string
	var clientType ssomanager.ClientType
	err := result.Scan(&clientName, &clientType)

	if err != nil {
		log.Errorf("Error while getting authorization client query: %v", err)
		return clientName, clientType, err
	}

	return clientName, clientType, nil

}
