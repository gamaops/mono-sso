package datastore

import (
	sso "github.com/gamaops/mono-sso/pkg/idl/sso-service"
)

func (d *Datastore) RegisterEvent(event *sso.RegisterEventRequest) {

	d.Logger.WithFields(map[string]interface{}{
		"level":        event.Level,
		"is_sensitive": event.IsSensitive,
		"data":         event.Data,
		"source":       "event",
	}).Info(event.Message)

}
