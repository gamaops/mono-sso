package datastore

import (
	sso "github.com/gamaops/mono-sso/pkg/idl/sso-service"
)

func (d *Datastore) RegisterEvent(event *sso.RegisterEventRequest) {

	fields := make(map[string]interface{}, len(event.Data)+3)
	fields["level"] = event.Level
	fields["is_sensitive"] = event.IsSensitive
	fields["source"] = "event"

	for key, value := range event.Data {
		fields[key] = value
	}

	d.Logger.WithFields(fields).Info(event.Message)

}
