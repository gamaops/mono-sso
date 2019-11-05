package datastore

import (
	"strings"

	"github.com/gamaops/mono-sso/pkg/constants"
	sso "github.com/gamaops/mono-sso/pkg/idl/sso-common"
	"gopkg.in/go-playground/validator.v9"
)

func ParseErrorIntoStatus(err error, status *sso.ResponseStatus) *sso.ResponseStatus {

	switch terr := err.(type) {
	case validator.ValidationErrors:
		for _, verr := range terr {
			var message strings.Builder
			message.WriteString("Field \"")
			message.WriteString(verr.Field())
			message.WriteString("\" failed on validation ")
			message.WriteString(verr.ActualTag())
			status.Errors = append(status.Errors, &sso.ResponseStatus_Error{
				Slug:    constants.InvalidRequestSlg,
				Message: message.String(),
			})
		}
		break
	case error:
		status.Errors = append(status.Errors, &sso.ResponseStatus_Error{
			Slug:    constants.InternalErrorSlg,
			Message: terr.Error(),
		})
	}

	return status
}
