package datastore

import (
	"context"

	"time"

	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"

	"golang.org/x/crypto/bcrypt"

	sso "github.com/gamaops/mono-sso/pkg/idl/sso-manager"
	"github.com/lib/pq"
)

type ClientDoc struct {
	ID           string
	Name         string `validate:"required" json:"name"`
	Secret       string
	RedirectURIs []string       `validate:"required,min=1,unique,dive,url" json:"redirect_uris"`
	Type         sso.ClientType `validate:"min=0,max=1" json:"type"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    time.Time
}

func (c *ClientDoc) FromUpsertClientRequest(req *sso.UpsertClientRequest) {
	c.RedirectURIs = req.RedirectUris
	c.Name = req.Name
	c.Type = req.Type
}

var ErrClientNotFound = errors.New("client not found")

func (d *Datastore) ClientExistsByID(ctx context.Context, clientID string, tx *sql.Tx) (bool, error) {
	query := `SELECT count(1) FROM sso.client WHERE id = $1`
	var count int
	var err error = nil
	if tx != nil {
		result := tx.QueryRowContext(ctx, query, clientID)
		err = result.Scan(&count)
	} else {
		result := d.Client.QueryRowContext(ctx, query, clientID)
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

func (d *Datastore) UpsertClient(ctx context.Context, req *sso.UpsertClientRequest) (*ClientDoc, error) {

	tx, err := d.Client.BeginTx(ctx, nil)
	if err != nil {
		d.Logger.Errorf("Error when starting transaction: %v", err)
		return nil, err
	}
	defer tx.Rollback()

	client := &ClientDoc{}

	opTimestamp := time.Unix(req.Session.Timestamp, 0)

	if len(req.ClientId) > 0 {
		result := tx.QueryRow(`SELECT updated_at FROM sso.client WHERE id = $1`, req.ClientId)
		err := result.Scan(&client.UpdatedAt)

		if err == sql.ErrNoRows {
			return nil, ErrClientNotFound
		} else if err != nil {
			d.Logger.Errorf("Error when getting client from database: %v", err)
			return nil, err
		}

		if client.UpdatedAt.Unix() >= opTimestamp.Unix() {
			return nil, ErrVersionMismatch
		}
		client.UpdatedAt = opTimestamp
		client.ID = req.ClientId
		client.FromUpsertClientRequest(req)
		err = d.Options.Validator.Struct(client)
		if err != nil {
			return nil, err
		}
		_, err = tx.Exec(
			`UPDATE sso.client
			SET name=$1, type=$2, redirect_uris=$3, updated_at=$4
			WHERE id=$5`,
			client.Name,
			client.Type,
			pq.Array(client.RedirectURIs),
			client.UpdatedAt,
			client.ID,
		)
		if err != nil {
			d.Logger.Errorf("Error when updating client: %v", err)
			return nil, err
		}
		err = tx.Commit()
		if err != nil {
			d.Logger.Errorf("Error when committing upsert client transaction: %v", err)
			return nil, err
		}
		return client, nil
	}

	client.FromUpsertClientRequest(req)

	err = d.Options.Validator.Struct(client)
	if err != nil {
		return nil, err
	}

	client.CreatedAt = opTimestamp
	client.UpdatedAt = opTimestamp

	clientID, err := d.IDGenerator28.New()
	if err != nil {
		d.Logger.Errorf("Error when generating new client ID: %v", err)
		return nil, err
	}
	client.ID = clientID.Base32()

	secret := make([]byte, 36)
	_, err = rand.Read(secret)
	if err != nil {
		d.Logger.Errorf("Error when generating client secret: %v", err)
		return nil, err
	}
	client.Secret = base64.StdEncoding.WithPadding(base64.NoPadding).EncodeToString(secret)
	secretHash, err := bcrypt.GenerateFromPassword(secret, 10)
	if err != nil {
		d.Logger.Errorf("Error when hashing (bcrypt) client secret: %v", err)
		return nil, err
	}

	_, err = tx.Exec(
		`INSERT INTO sso.client (id, type, name, secret, redirect_uris, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		client.ID,
		client.Type,
		client.Name,
		secretHash,
		pq.Array(client.RedirectURIs),
		client.CreatedAt,
		client.UpdatedAt,
	)
	if err != nil {
		d.Logger.Errorf("Error when inserting client: %v", err)
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		d.Logger.Errorf("Error when committing upsert client transaction: %v", err)
		return nil, err
	}

	return client, nil

}
