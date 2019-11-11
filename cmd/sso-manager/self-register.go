package main

import (
	"context"
	"time"

	ssocommon "github.com/gamaops/mono-sso/pkg/idl/sso-common"
	sso "github.com/gamaops/mono-sso/pkg/idl/sso-manager"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

func registerSSOManagerApp(ctx context.Context) error {

	result := ServiceDatastore.Client.QueryRowContext(ctx, `SELECT nextval('sso.setup_admin')`)

	var currentSetupSeq int
	err := result.Scan(&currentSetupSeq)

	if err != nil {
		log.Warnf("Error when getting the sequence to start the admin setup, maybe it can be ignored: %v", err)
		return nil
	}

	if currentSetupSeq != 2 {
		log.Errorf("Invalid sequence to setup admin: %v", currentSetupSeq)
		return err
	}

	now := time.Now()

	clientReq := &sso.UpsertClientRequest{
		Session: &ssocommon.RequestSession{
			Timestamp: now.Unix(),
		},
		Type:         sso.ClientType_CONFIDENTIAL,
		Name:         viper.GetString("clientAppName"),
		RedirectUris: viper.GetStringSlice("clientAppRedirectUris"),
	}

	client, err := ServiceDatastore.UpsertClient(ctx, clientReq)

	if err != nil {
		log.Errorf("Error when setting up admin: %v", err)
		return nil
	}

	log.Warnf("Your SSO Manager client app was set up, the secret is \"%v\" and the ID is \"%v\"", client.Secret, client.ID)

	tx, err := ServiceDatastore.Client.BeginTx(ctx, nil)
	if err != nil {
		log.Errorf("Error when starting transaction to create admin account: %v", err)
		return err
	}
	defer tx.Rollback()

	accID, err := ServiceDatastore.IDGenerator28.New()
	if err != nil {
		log.Errorf("Error when generating admin account ID: %v", err)
		return err
	}
	accIDStr := accID.Base32()

	_, err = tx.Exec(
		`INSERT INTO sso.tenant (id, created_at, updated_at, name)
		VALUES ($1, $2, $3, 'SSO Administration')`,
		viper.GetString("adminTenant"),
		now,
		now,
	)
	if err != nil {
		log.Errorf("Error when inserting admin tenant: %v", err)
		return err
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(viper.GetString("adminAccountPassword")),
		10,
	)
	if err != nil {
		log.Errorf("Error when hashing admin password: %v", err)
		return err
	}

	_, err = tx.Exec(
		`INSERT INTO sso.account (id, name, activation_method, password, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		accIDStr,
		viper.GetString("adminAccountName"),
		0,
		passwordHash,
		now,
		now,
	)
	if err != nil {
		log.Errorf("Error when inserting admin account: %v", err)
		return err
	}

	_, err = tx.Exec(
		`INSERT INTO sso.account_identifier (account_id, identifier, created_at)
		VALUES ($1, $2, $3)`,
		accIDStr,
		viper.GetString("adminAccountIdentifier"),
		now,
	)
	if err != nil {
		log.Errorf("Error when inserting admin account identifier: %v", err)
		return err
	}

	_, err = tx.Exec(
		`INSERT INTO sso.account_tenant (account_id, tenant_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4)`,
		accIDStr,
		viper.GetString("adminTenant"),
		now,
		now,
	)
	if err != nil {
		log.Errorf("Error when inserting admin account and tenant relation: %v", err)
		return err
	}

	scopeID, err := ServiceDatastore.IDGenerator28.New()
	if err != nil {
		log.Errorf("Error when generating superadmin scope ID: %v", err)
		return err
	}
	scopeIDStr := scopeID.Base32()

	_, err = tx.Exec(
		`INSERT INTO sso.scope (id, client_id, created_at, updated_at, scope)
		VALUES ($1, $2, $3, $4, 'superadmin')`,
		scopeIDStr,
		client.ID,
		now,
		now,
	)
	if err != nil {
		log.Errorf("Error when inserting superadmin scope: %v", err)
		return err
	}

	_, err = tx.Exec(
		`INSERT INTO sso.scope_i18n (scope_id, created_at, updated_at, locale, description)
		VALUES ($1, $2, $3, 'en-US', 'Superadmin role')`,
		scopeIDStr,
		now,
		now,
	)
	if err != nil {
		log.Errorf("Error when inserting superadmin scope i18n: %v", err)
		return err
	}

	err = tx.Commit()

	return nil

}
