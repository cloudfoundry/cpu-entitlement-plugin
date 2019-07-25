package token

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

//go:generate counterfeiter . GetToken
type GetToken func() (string, error)

type Getter struct {
	getToken            GetToken
	currentToken        string
	tokenExpirationTime time.Time
}

func NewGetter(getToken GetToken) *Getter {
	return &Getter{
		getToken:            getToken,
		tokenExpirationTime: time.Now(),
	}
}

func (t *Getter) Token() (string, error) {
	if t.tokenExpired() {
		return t.refreshToken()
	}
	return t.currentToken, nil
}

func (t *Getter) tokenExpired() bool {
	return time.Now().After(t.tokenExpirationTime)
}

func (t *Getter) refreshToken() (string, error) {
	token, err := t.getToken()
	if err != nil {
		return "", err
	}

	t.currentToken = token

	t.tokenExpirationTime, err = extractExpirationTimeFromToken(token)
	if err != nil {
		return "", err
	}
	return token, nil
}

type tokenJson struct {
	ExpTime int64 `json:"exp"`
}

func extractExpirationTimeFromToken(token string) (time.Time, error) {
	encodedMetadata := strings.Split(token, ".")[1]

	decodedMetadata, err := base64.RawURLEncoding.DecodeString(encodedMetadata)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to decode token from base64: %s", err.Error())
	}

	var metadata tokenJson
	err = json.Unmarshal(decodedMetadata, &metadata)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid token: %s", err.Error())
	}

	return time.Unix(metadata.ExpTime, 0), nil
}
