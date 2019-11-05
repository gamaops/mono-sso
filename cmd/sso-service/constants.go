package main

import (
	"github.com/gamaops/mono-sso/pkg/constants"
	sso "github.com/gamaops/mono-sso/pkg/idl/sso-common"
)

var SignInInvalidAccountStatus = &sso.ResponseStatus{
	Errors: []*sso.ResponseStatus_Error{
		&sso.ResponseStatus_Error{
			Slug:    constants.InvalidAccountSlg,
			Message: constants.InvalidAccountMsg,
		},
	},
}

var InvalidActivationRequestStatus = &sso.ResponseStatus{
	Errors: []*sso.ResponseStatus_Error{
		&sso.ResponseStatus_Error{
			Slug:    constants.InvalidActivationSlg,
			Message: constants.InvalidActivationMsg,
		},
	},
}

var InternalErrorStatus = &sso.ResponseStatus{
	Errors: []*sso.ResponseStatus_Error{
		&sso.ResponseStatus_Error{
			Slug:    constants.InternalErrorSlg,
			Message: constants.InternalErrorMsg,
		},
	},
}

var UnknownScopesStatus = &sso.ResponseStatus{
	Errors: []*sso.ResponseStatus_Error{
		&sso.ResponseStatus_Error{
			Slug:    constants.UknownScopesSlg,
			Message: constants.UknownScopesMsg,
		},
	},
}

var InvalidClientStatus = &sso.ResponseStatus{
	Errors: []*sso.ResponseStatus_Error{
		&sso.ResponseStatus_Error{
			Slug:    constants.InvalidClientSlg,
			Message: constants.InvalidClientMsg,
		},
	},
}

var InvalidResponseTypeStatus = &sso.ResponseStatus{
	Errors: []*sso.ResponseStatus_Error{
		&sso.ResponseStatus_Error{
			Slug:    constants.InvalidResponseTypeSlg,
			Message: constants.InvalidResponseTypeMsg,
		},
	},
}

var NoValidRefreshTokenStatus = &sso.ResponseStatus{
	Errors: []*sso.ResponseStatus_Error{
		&sso.ResponseStatus_Error{
			Slug:    constants.NoValidRefreshTokenSlg,
			Message: constants.NoValidRefreshTokenMsg,
		},
	},
}

var InvalidGrantStatus = &sso.ResponseStatus{
	Errors: []*sso.ResponseStatus_Error{
		&sso.ResponseStatus_Error{
			Slug:    constants.InvalidGrantSlg,
			Message: constants.InvalidGrantMsg,
		},
	},
}

var UnauthorizedScopesStatus = &sso.ResponseStatus{
	Errors: []*sso.ResponseStatus_Error{
		&sso.ResponseStatus_Error{
			Slug:    constants.UnauthorizedScopesSlg,
			Message: constants.UnauthorizedScopesMsg,
		},
	},
}
