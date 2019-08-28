package token

import (
	"net/http"
)

func AuthenticatedBy(tokenGetter *Getter) *AuthClient {
	return &AuthClient{tokenGetter: tokenGetter}
}

type AuthClient struct {
	tokenGetter *Getter
}

func (a *AuthClient) Do(req *http.Request) (*http.Response, error) {
	t, err := a.tokenGetter.Token()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", t)
	return http.DefaultClient.Do(req)
}
