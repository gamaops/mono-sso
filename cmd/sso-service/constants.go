package main

import (
	"github.com/gamaops/mono-sso/pkg/constants"
	sso "github.com/gamaops/mono-sso/pkg/idl/sso-common"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
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

var ClientAuthorizationOpts = options.FindOne().SetProjection(bson.D{
	{"name", 1},
	{"type", 1},
})

var SignInFindOpts = options.FindOne().SetProjection(bson.D{
	{"_id", 1},
	{"activation_method", 1},
	{"name", 1},
	{"password", 1},
})

var FindRefreshTokenOpts = options.FindOne().SetProjection(bson.D{
	{"_id", 1},
	{"expires_at", 1},
})

var FindClientRefreshTokenOpts = options.FindOne().SetProjection(bson.D{
	{"secret", 1},
})

var GrantUpdateOptions = options.Update().SetUpsert(true)
