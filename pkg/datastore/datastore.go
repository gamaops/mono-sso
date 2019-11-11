package datastore

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"errors"

	"github.com/gamaops/gamago/pkg/id"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"gopkg.in/go-playground/validator.v9"
)

var ErrVersionMismatch = errors.New("version mismatch")

type Options struct {
	PostgresURI    string
	MaxConnections int
	Validator      *validator.Validate
}

type Datastore struct {
	Options       *Options
	Client        *sql.DB
	Logger        *logrus.Logger
	IDGenerator28 *id.IDGenerator
}

func StartDatastore(ctx context.Context, datastore *Datastore) error {
	datastore.Logger.Info("Connecting to Postgres")
	db, err := sql.Open("postgres", datastore.Options.PostgresURI)
	if err != nil {
		return err
	}

	db.SetMaxOpenConns(datastore.Options.MaxConnections)

	datastore.Client = db

	datastore.IDGenerator28, err = id.NewIDGenerator(5)
	if err != nil {
		return err
	}

	return nil
}

func StopDataStore(ctx context.Context, datastore *Datastore) {
	datastore.Logger.Warn("Disconnecting from Postgres")
	if err := datastore.Client.Close(); err != nil {
		datastore.Logger.Errorf("Error while disconnecting from Postgres: %v", err)
	}
}

func CreatePlaceholders(start int, count int, builder *strings.Builder) *strings.Builder {
	if builder == nil {
		builder = &strings.Builder{}
	}
	for i := 1; i <= count; i++ {
		if i > 1 {
			builder.WriteRune(',')
		}
		builder.WriteRune('$')
		builder.WriteString(strconv.FormatInt(int64(i+start), 10))
	}
	return builder
}
