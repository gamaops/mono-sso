package datastore

import (
	"context"
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

func (d *Datastore) IsAccountInTenant(ctx context.Context, accountID string, tenantID string) (bool, error) {

	result := d.Client.QueryRowContext(
		ctx,
		"SELECT count(1) FROM sso.account_tenant WHERE account_id = $1 AND tenant_id = $2 AND deleted_at IS NULL",
		accountID,
		tenantID,
	)

	var count int
	err := result.Scan(&count)

	if err != nil {
		d.Logger.Errorf("Error when getting tenant count for account: %v", err)
		return false, err
	}

	if count > 0 {
		return true, nil
	}

	return false, nil

}

func (d *Datastore) RevokeScopes(ctx context.Context, req *sso.RevokeScopesRequest) error {

	tx, err := d.Client.BeginTx(ctx, nil)
	if err != nil {
		d.Logger.Errorf("Error when starting transaction: %v", err)
		return err
	}
	defer tx.Rollback()

	scopesHolder := CreatePlaceholders(1, len(req.Scopes), nil)
	args := []interface{}{
		req.ClientId,
	}
	for _, scope := range req.Scopes {
		args = append(args, scope)
	}

	result, err := tx.Query(`SELECT id FROM sso.scope WHERE client_id = $1 AND scopes IN (`+scopesHolder.String()+`)`, args...)
	if err != nil {
		d.Logger.Errorf("Error when selecting scopes to revoke: %v", err)
		return err
	}
	defer result.Close()

	args = []interface{}{
		req.Subject,
	}

	for result.Next() {
		var scopeID string
		if err := result.Scan(&scopeID); err != nil {
			d.Logger.Errorf("Error when iterating over scopes to revoke: %v", err)
			return err
		}
		args = append(args, scopeID)
	}

	_, err = tx.Exec(`UPDATE sso.grant SET deleted_at = now() WHERE deleted_at IS NULL AND account_id = $1 AND scope_id IN (`+scopesHolder.String()+`)`, args...)
	if err != nil {
		d.Logger.Errorf("Error when revoking scopes: %v", err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		d.Logger.Errorf("Error when committing transaction to revoke scopes: %v", err)
		return err
	}

	return nil

}

func (d *Datastore) RevokeToken(ctx context.Context, req *sso.RevokeTokenRequest) error {

	_, err := d.Client.Exec(`UPDATE sso.token SET deleted_at = now() WHERE deleted_at IS NULL AND account_id = $1 AND client_id = $2`, req.Subject, req.ClientId)
	if err != nil {
		d.Logger.Errorf("Error when revoking token: %v", err)
		return err
	}

	return nil

}
