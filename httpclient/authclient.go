package httpclient

import (
	"crypto/tls"
	"net/http"
)

type AuthClient struct {
	tokenGetter *TokenGetter
}

func NewAuthClient(getToken GetToken) *AuthClient {
	tokenGetter := NewTokenGetter(getToken)

	return &AuthClient{
		tokenGetter: tokenGetter,
	}
}

func (a *AuthClient) SkipSSLValidation() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
}

func (a *AuthClient) Do(req *http.Request) (*http.Response, error) {
	t, err := a.tokenGetter.Token()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", t)
	return http.DefaultClient.Do(req)
}
