package datastore

import (
	"context"
	"time"

	sso "github.com/gamaops/mono-sso/pkg/idl/sso-service"
	"go.mongodb.org/mongo-driver/bson"
)

func (d *Datastore) InsertEvent(ctx context.Context, event *sso.RegisterEventRequest) error {

	_, err := d.Collections.Events.InsertOne(ctx, bson.M{
		"level":        event.Level,
		"is_sensitive": event.IsSensitive,
		"message":      event.Message,
		"data":         event.Data,
		"created_at":   time.Now(),
	})

	return err

}
