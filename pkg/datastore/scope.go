package datastore

import (
	"context"
	"database/sql"
	"errors"
	"time"

	sso "github.com/gamaops/mono-sso/pkg/idl/sso-manager"
)

type ScopeDoc struct {
	ID        string
	ClientID  string `validate:"required,len=28" json:"client_id"`
	Scope     string `validate:"required,min=1,max=100" json:"scope"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

func (s *ScopeDoc) FromUpsertScopeRequest(req *sso.UpsertScopeRequest) {
	s.ClientID = req.ClientId
	s.Scope = req.Scope
}

type ScopeI18nDoc struct {
	Description string `validate:"required,min=1,max=200" json:"description"`
	Locale      string
}

func (s *ScopeI18nDoc) FromUpsertScopeRequest(req *sso.UpsertScopeRequest) {
	s.Locale = req.Session.Locale
	s.Description = req.Description
}

var ErrScopeNotFound = errors.New("scope not found")

func (d *Datastore) UpsertScope(ctx context.Context, req *sso.UpsertScopeRequest) (*ScopeDoc, error) {

	tx, err := d.Client.BeginTx(ctx, nil)
	if err != nil {
		d.Logger.Errorf("Error when starting transaction: %v", err)
		return nil, err
	}
	defer tx.Rollback()

	scope := &ScopeDoc{}
	scopeI18n := &ScopeI18nDoc{}

	opTimestamp := time.Unix(req.Session.Timestamp, 0)

	result := tx.QueryRow(
		`SELECT sco.updated_at, sco.id
		FROM sso.scope AS sco
		WHERE sco.scope = $1 AND sco.client_id = $2`,
		req.Scope,
		req.ClientId,
	)
	err = result.Scan(&scope.UpdatedAt, &scope.ID)

	if err == sql.ErrNoRows {
		scope.UpdatedAt = opTimestamp
		scope.CreatedAt = opTimestamp
		scopeID, err := d.IDGenerator28.New()
		if err != nil {
			d.Logger.Errorf("Error when generating new scope ID: %v", err)
			return nil, err
		}
		scope.ID = scopeID.Base32()
	} else if err != nil {
		d.Logger.Errorf("Error when getting scope from database: %v", err)
		return nil, err
	} else if scope.UpdatedAt.Unix() >= opTimestamp.Unix() {
		return nil, ErrVersionMismatch
	}

	exists, err := d.ClientExistsByID(ctx, req.ClientId, tx)

	if err != nil {
		return nil, err
	} else if !exists {
		return nil, ErrClientNotFound
	}

	scope.UpdatedAt = opTimestamp

	scope.FromUpsertScopeRequest(req)
	err = d.Options.Validator.Struct(scope)
	if err != nil {
		return nil, err
	}

	scopeI18n.FromUpsertScopeRequest(req)
	err = d.Options.Validator.Struct(scopeI18n)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(
		`INSERT INTO sso.scope (id, client_id, scope, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT ON CONSTRAINT unq_sso_scope_id DO UPDATE SET updated_at = $5`,
		scope.ID,
		scope.ClientID,
		scope.Scope,
		opTimestamp,
		opTimestamp,
	)
	if err != nil {
		d.Logger.Errorf("Error when updating scope: %v", err)
		return nil, err
	}

	_, err = tx.Exec(
		`INSERT INTO sso.scope_i18n (scope_id, locale, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT ON CONSTRAINT unq_sso_scope_i18n_scope_id_locale DO UPDATE SET description = $3, updated_at = $5`,
		scope.ID,
		scopeI18n.Locale,
		scopeI18n.Description,
		opTimestamp,
		opTimestamp,
	)
	if err != nil {
		d.Logger.Errorf("Error when updating scope i18n: %v", err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		d.Logger.Errorf("Error when committing upsert scope transaction: %v", err)
		return nil, err
	}
	return scope, nil

}
