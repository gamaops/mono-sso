package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	sso "github.com/gamaops/mono-sso/pkg/idl/sso-service"
)

func hashActivationCode(code []byte, userAgent []byte) string {
	hash := sha256.New()
	hash.Write(code)
	hash.Write(userAgent)
	return hex.EncodeToString(hash.Sum(nil))
}

func newActivationCode(userAgent []byte) (string, string) {
	min := 1
	max := 999999
	code := fmt.Sprintf("%06d", rand.Intn(max-min+1)+min)
	return code, hashActivationCode([]byte(code), userAgent)
}

func generateActivationCode(ctx context.Context, acc *AccountEntity, signInReq *sso.SignInRequest) error {

	if acc.ActivationMethod == sso.ActivationMethod_NONE {
		return nil
	}

	if acc.ActivationMethod == sso.ActivationMethod_GOOGLE_AUTHENTICATOR {
		// TODO: Don't need to generate a new activation code
		return nil
	}

	code, hash := newActivationCode([]byte(signInReq.UserAgent))
	subject := acc.ID.Hex()

	err := ServiceDatastore.InsertEvent(ctx, &sso.RegisterEventRequest{
		Level:       sso.EventLevel_INFO,
		IsSensitive: true,
		Message:     fmt.Sprintf("new activation code (MFA enabled): %v", subject),
		Data: map[string]string{
			"account_id":        subject,
			"user_agent":        signInReq.UserAgent,
			"session_id":        signInReq.SessionId,
			"activation_mehtod": strconv.FormatInt(int64(acc.ActivationMethod), 10),
			"activation_code":   code,
			"duration":          signInReq.ActivationCodeDuration,
		},
	})

	if err != nil {
		log.Errorf("Error when inserting event: %v", err)
		return err
	}

	expiration, err := time.ParseDuration(signInReq.ActivationCodeDuration)

	if err != nil {
		log.Errorf("Error when parsing activation code duration: %v", err)
		return err
	}

	actCodeKey, err := ServiceCache.CreateID(":act:", signInReq.SessionId, ':', hash)
	if err != nil {
		log.Errorf("Error when generating activation code cache ID: %v", err)
		return err
	}

	return ServiceCache.Client.Set(actCodeKey.String(), subject, expiration).Err()

}

func validateActivationCode(ctx context.Context, actReq *sso.ActivateSessionRequest) (bool, error) {

	hash := hashActivationCode([]byte(actReq.ActivationCode), []byte(actReq.UserAgent))

	actCodeKey, err := ServiceCache.CreateID(":act:", actReq.SessionId, ':', hash)
	if err != nil {
		log.Errorf("Error when generating activation code cache ID: %v", err)
		return false, err
	}

	key := actCodeKey.String()

	pipe := ServiceCache.Client.TxPipeline()

	get := pipe.Get(key)
	pipe.Del(key)

	_, err = pipe.Exec()

	if err != nil {
		log.Fatalf("Error while executing activation session Redis transaction: %v", err)
		return false, nil
	}

	subject, err := get.Result()

	if err != nil {
		log.Warnf("Error when getting result from activation code validation: %v", err)
		err = ServiceDatastore.InsertEvent(ctx, &sso.RegisterEventRequest{
			Level:       sso.EventLevel_WARNING,
			IsSensitive: false,
			Message:     fmt.Sprintf("failed to activate session (invalid code/user agent/session): %v", subject),
			Data: map[string]string{
				"account_id": subject,
				"user_agent": actReq.UserAgent,
				"session_id": actReq.SessionId,
			},
		})
		return false, err
	}

	if subject != actReq.Subject {
		err = ServiceDatastore.InsertEvent(ctx, &sso.RegisterEventRequest{
			Level:       sso.EventLevel_WARNING,
			IsSensitive: false,
			Message:     fmt.Sprintf("failed to activate session (invalid subject): %v", subject),
			Data: map[string]string{
				"account_id": subject,
				"user_agent": actReq.UserAgent,
				"session_id": actReq.SessionId,
			},
		})
		return false, err
	}

	err = ServiceDatastore.InsertEvent(ctx, &sso.RegisterEventRequest{
		Level:       sso.EventLevel_INFO,
		IsSensitive: true,
		Message:     fmt.Sprintf("session activated: %v", subject),
		Data: map[string]string{
			"account_id": subject,
			"user_agent": actReq.UserAgent,
			"session_id": actReq.SessionId,
			"code":       actReq.ActivationCode,
		},
	})

	return true, err

}
