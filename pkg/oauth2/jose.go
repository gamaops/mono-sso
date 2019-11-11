package oauth2

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
	jose "github.com/square/go-jose/v3"
	"github.com/square/go-jose/v3/jwt"
)

var ErrInvalidFingerprint error = errors.New("invalid key fingerprint")
var ErrInvalidKeyType error = errors.New("invalid key type, must be RSA or EC")
var ErrEmptyJWKS error = errors.New("JWKS is empty")

type Options struct {
	PrivateKeyPath     string
	PrivateKeyPassword string
	JWKSURL            string
}

type JWKS struct {
	Keys []jose.JSONWebKey `json:"keys"`
}

type OAuth2Jose struct {
	Options     *Options
	Fingerprint string
	RSAKey      *rsa.PrivateKey
	ECDSAKey    *ecdsa.PrivateKey
	JWK         *jose.JSONWebKey
	JWKS        []byte
	SigningKey  *jose.SigningKey
	Signer      jose.Signer
	Logger      *logrus.Logger
}

type TokenClaims struct {
	*jwt.Claims
	Scope  string `json:"scope,omitempty"`
	Kid    string `json:"kid,omitempty"`
	Tenant string `json:"tenant,omitempty"`
}

func SetupOAuth2Jose(oauth2jose *OAuth2Jose) error {

	privateKeyContent, err := ioutil.ReadFile(oauth2jose.Options.PrivateKeyPath)

	if err != nil {
		oauth2jose.Logger.Warn("No private key found, generating temporary one")
		privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			return err
		}
		privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
		fingerprint := sha1.Sum(privateKeyBytes)
		oauth2jose.Fingerprint = hex.EncodeToString(fingerprint[:])
		oauth2jose.RSAKey = privateKey
		return setupJWK(oauth2jose)
	}

	privatePem, _ := pem.Decode(privateKeyContent)
	var privatePemBytes []byte
	if privatePem.Type != "RSA PRIVATE KEY" && privatePem.Type != "EC PRIVATE KEY" {
		return ErrInvalidKeyType
	}

	privatePemPassword := []byte(oauth2jose.Options.PrivateKeyPassword)

	if len(privatePemPassword) > 0 {
		privatePemBytes, err = x509.DecryptPEMBlock(privatePem, privatePemPassword)
		if err != nil {
			return err
		}
	} else {
		privatePemBytes = privatePem.Bytes
	}

	fingerprint := sha1.Sum(privatePemBytes)
	oauth2jose.Fingerprint = hex.EncodeToString(fingerprint[:])

	if privatePem.Type == "RSA PRIVATE KEY" {
		var parsedKey interface{}
		if parsedKey, err = x509.ParsePKCS1PrivateKey(privatePemBytes); err != nil {
			if parsedKey, err = x509.ParsePKCS8PrivateKey(privatePemBytes); err != nil {
				return err
			}
		}

		var privateKey *rsa.PrivateKey
		privateKey, _ = parsedKey.(*rsa.PrivateKey)
		oauth2jose.RSAKey = privateKey
	} else if privatePem.Type == "EC PRIVATE KEY" {
		var privateKey *ecdsa.PrivateKey
		if privateKey, err = x509.ParseECPrivateKey(privatePemBytes); err != nil {
			return err
		}
		oauth2jose.ECDSAKey = privateKey
	}

	return setupJWK(oauth2jose)

}

func setupJWK(oauth2jose *OAuth2Jose) error {
	oauth2jose.JWK = &jose.JSONWebKey{
		KeyID: oauth2jose.Fingerprint,
		Use:   "sig",
	}

	oauth2jose.SigningKey = &jose.SigningKey{}

	if oauth2jose.RSAKey != nil {
		oauth2jose.JWK.Key = oauth2jose.RSAKey.Public()
		oauth2jose.JWK.Algorithm = "RS256"
		oauth2jose.SigningKey.Key = oauth2jose.RSAKey
		oauth2jose.SigningKey.Algorithm = jose.RS256
	} else {
		oauth2jose.JWK.Key = oauth2jose.ECDSAKey.Public()
		oauth2jose.JWK.Algorithm = "ES256"
		oauth2jose.SigningKey.Key = oauth2jose.ECDSAKey
		oauth2jose.SigningKey.Algorithm = jose.ES256
	}

	jwkJSON, _ := oauth2jose.JWK.MarshalJSON()

	jwks := &bytes.Buffer{}

	jwks.WriteString("{\"keys\":[")
	jwks.Write(jwkJSON)
	jwks.WriteString("]}")

	oauth2jose.JWKS = jwks.Bytes()

	signer, err := jose.NewSigner(*(oauth2jose.SigningKey), (&jose.SignerOptions{}).WithType("JWT"))
	if err != nil {
		return err
	}

	oauth2jose.Signer = signer

	return nil
}

func GenerateJWT(oauth2jose *OAuth2Jose, claims *TokenClaims) jwt.Builder {

	claims.Kid = oauth2jose.Fingerprint
	builder := jwt.Signed(oauth2jose.Signer)
	builder = builder.Claims(claims)

	return builder

}

func DecodeJWT(oauth2jose *OAuth2Jose, tokenStr string) (*TokenClaims, error) {
	token, err := jwt.ParseSigned(tokenStr)

	if err != nil {
		return nil, err
	}

	claims := &TokenClaims{
		Claims: &jwt.Claims{},
	}
	err = token.Claims(oauth2jose.JWK, claims)
	if err != nil {
		return nil, err
	}

	if claims.Kid != oauth2jose.Fingerprint {
		return nil, ErrInvalidFingerprint
	}

	return claims, nil

}

func LoadOAuth2JoseFromURL(oauth2jose *OAuth2Jose) error {

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	defer client.CloseIdleConnections()

	res, err := client.Get(oauth2jose.Options.JWKSURL)

	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	jwks := new(JWKS)
	err = json.Unmarshal(body, jwks)
	if err != nil {
		return err
	}

	if len(jwks.Keys) == 0 {
		return ErrEmptyJWKS
	}

	oauth2jose.JWK = &jwks.Keys[0]
	oauth2jose.Fingerprint = oauth2jose.JWK.KeyID

	return nil
}
