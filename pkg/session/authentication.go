package session

import (
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/gamaops/mono-sso/pkg/cache"
	sso "github.com/gamaops/mono-sso/pkg/idl/sso-service"
	"github.com/go-redis/redis"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
)

type AuthenticationOptions struct {
	IndexTemplatePath        string
	Namespace                string
	RememberMeDuration       time.Duration
	EphemeralSessionDuration time.Duration
	MFASessionDuration       time.Duration
	SessionCookieDomain      string
	SessionCookiePath        string
	AccountServiceClient     sso.AccountServiceClient
}

type AuthenticationModel struct {
	Options          *AuthenticationOptions
	SessionCookieKey string
	SubjectCookieKey string
	IndexTemplate    *template.Template
	Logger           *logrus.Logger
}

func SetupAuthenticationModel(model *AuthenticationModel) error {
	key := &strings.Builder{}
	key.WriteString(model.Options.Namespace)
	key.WriteString("_SESS")
	model.SessionCookieKey = key.String()

	key = &strings.Builder{}
	key.WriteString(model.Options.Namespace)
	key.WriteString("_SUB")
	model.SubjectCookieKey = key.String()

	var err error = nil
	model.IndexTemplate, err = template.ParseFiles(model.Options.IndexTemplatePath)
	return err
}

func GetCachedSessionSubject(chc *cache.Cache, sessSubID string) (*sso.SessionSubject, error) {
	currentSess := chc.Client.Get(sessSubID)
	currentSessProto, err := currentSess.Bytes()
	var sessSub *sso.SessionSubject = nil
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		chc.Logger.Errorf("Error when getting session from Redis: %v", err)
		return nil, err
	}
	sessSub = new(sso.SessionSubject)
	err = proto.Unmarshal(currentSessProto, sessSub)
	if err != nil {
		chc.Logger.Errorf("Error when unmarshaling session subject: %v", err)
		return nil, err
	}
	return sessSub, nil
}

func GetExpirationFromAuthentication(model *AuthenticationModel, authReq *AuthenticationRequest) (time.Duration, bool) {
	if model.Options.RememberMeDuration > 0 && authReq.RememberMe {
		return model.Options.RememberMeDuration, true
	}
	return model.Options.EphemeralSessionDuration, false
}

func UpdateSessionSubject(chc *cache.Cache, sessSubID string, sessSub *sso.SessionSubject, expiration time.Duration) error {
	data, err := proto.Marshal(sessSub)
	if err != nil {
		chc.Logger.Errorf("Error when encoding session subject: %v", err)
		return err
	}
	res := chc.Client.Set(sessSubID, data, expiration)
	err = res.Err()
	if err != nil {
		chc.Logger.Errorf("Error saving session on Redis: %v", err)
		return err
	}
	return nil
}

func (m *AuthenticationModel) GetSessionAndSubjectCookies(r *http.Request) (*http.Cookie, *http.Cookie, error) {
	subjectCookie, err := r.Cookie(m.SubjectCookieKey) // TODO: Accept subject from URL
	if err != nil && err != http.ErrNoCookie {
		return nil, nil, err
	}

	sessionCookie, err := r.Cookie(m.SessionCookieKey)
	if err != nil && err != http.ErrNoCookie {
		return nil, nil, err
	}

	return sessionCookie, subjectCookie, nil
}
