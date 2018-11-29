package token

import (
	"sync"
	"time"
)

//go:generate counterfeiter . GetToken
type GetToken func() (string, error)

type TokenGetter struct {
	getToken     GetToken
	expired      bool
	currentToken string
	mutex        *sync.Mutex
}

func NewTokenGetter(getToken GetToken, tokenLifetime time.Duration) *TokenGetter {
	ticker := time.NewTicker(tokenLifetime)
	tokenGetter := &TokenGetter{
		getToken: getToken,
		expired:  true,
		mutex:    &sync.Mutex{},
	}
	go tokenGetter.expireToken(ticker)

	return tokenGetter
}

func (t *TokenGetter) Token() (string, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.expired {
		return t.refreshToken()
	}
	return t.currentToken, nil
}

func (t *TokenGetter) expireToken(ticker *time.Ticker) {
	for range ticker.C {
		t.mutex.Lock()
		t.expired = true
		t.mutex.Unlock()
	}
}

func (t *TokenGetter) refreshToken() (string, error) {
	token, err := t.getToken()
	if err != nil {
		return "", err
	}

	t.currentToken = token
	t.expired = false
	return token, nil
}
