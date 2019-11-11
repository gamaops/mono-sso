package datastore

import (
	"context"
	"database/sql"
	"errors"
	"time"

	sso "github.com/gamaops/mono-sso/pkg/idl/sso-manager"
)

type TenantDoc struct {
	ID        string
	Name      string `validate:"required,min=1,max=60" json:"name"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

func (c *TenantDoc) FromUpsertTenantRequest(req *sso.UpsertTenantRequest) {
	c.Name = req.Name
}

var ErrTenantNotFound = errors.New("tenant not found")

func (d *Datastore) TenantExistsByID(ctx context.Context, tenantID string, tx *sql.Tx) (bool, error) {
	query := `SELECT count(1) FROM sso.tenant WHERE id = $1`
	var count int
	var err error = nil
	if tx != nil {
		result := tx.QueryRowContext(ctx, query, tenantID)
		err = result.Scan(&count)
	} else {
		result := d.Client.QueryRowContext(ctx, query, tenantID)
		err = result.Scan(&count)
	}
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}

func (d *Datastore) UpsertTenant(ctx context.Context, req *sso.UpsertTenantRequest) (*TenantDoc, error) {

	tx, err := d.Client.BeginTx(ctx, nil)
	if err != nil {
		d.Logger.Errorf("Error when starting transaction: %v", err)
		return nil, err
	}
	defer tx.Rollback()

	tenant := &TenantDoc{}

	opTimestamp := time.Unix(req.Session.Timestamp, 0)

	if len(req.TenantId) > 0 {
		result := tx.QueryRow(`SELECT updated_at FROM sso.tenant WHERE id = $1`, req.TenantId)
		err := result.Scan(&tenant.UpdatedAt)

		if err == sql.ErrNoRows {
			return nil, ErrTenantNotFound
		} else if err != nil {
			d.Logger.Errorf("Error when getting tenant from database: %v", err)
			return nil, err
		}

		if tenant.UpdatedAt.Unix() >= opTimestamp.Unix() {
			return nil, ErrVersionMismatch
		}
		tenant.UpdatedAt = opTimestamp
		tenant.ID = req.TenantId
		tenant.FromUpsertTenantRequest(req)
		err = d.Options.Validator.Struct(tenant)
		if err != nil {
			return nil, err
		}
		_, err = tx.Exec(
			`UPDATE sso.tenant
			SET name=$1, updated_at=$2
			WHERE id=$3`,
			tenant.Name,
			tenant.UpdatedAt,
			tenant.ID,
		)
		if err != nil {
			d.Logger.Errorf("Error when updating tenant: %v", err)
			return nil, err
		}
		err = tx.Commit()
		if err != nil {
			d.Logger.Errorf("Error when committing upsert tenant transaction: %v", err)
			return nil, err
		}
		return tenant, nil
	}

	tenant.FromUpsertTenantRequest(req)

	err = d.Options.Validator.Struct(tenant)
	if err != nil {
		return nil, err
	}

	tenant.CreatedAt = opTimestamp
	tenant.UpdatedAt = opTimestamp

	tenantID, err := d.IDGenerator28.New()
	if err != nil {
		d.Logger.Errorf("Error when generating new tenant ID: %v", err)
		return nil, err
	}
	tenant.ID = tenantID.Base32()

	_, err = tx.Exec(
		`INSERT INTO sso.tenant (id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4)`,
		tenant.ID,
		tenant.Name,
		tenant.CreatedAt,
		tenant.UpdatedAt,
	)
	if err != nil {
		d.Logger.Errorf("Error when inserting tenant: %v", err)
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		d.Logger.Errorf("Error when committing upsert tenant transaction: %v", err)
		return nil, err
	}

	return tenant, nil

}
