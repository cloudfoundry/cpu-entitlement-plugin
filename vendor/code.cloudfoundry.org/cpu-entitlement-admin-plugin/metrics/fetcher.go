package metrics

import (
	"context"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cpu-entitlement-plugin/token"
	logcache "code.cloudfoundry.org/log-cache/pkg/client"
)

type LogCacheFetcher struct {
	logCacheClient *logcache.Client
}

func NewLogCacheFetcher(logCacheURL string, getToken token.GetToken) LogCacheFetcher {
	logCacheClient := logcache.NewClient(
		logCacheURL,
		logcache.WithHTTPClient(authenticatedBy(token.NewGetter(getToken))),
	)

	return LogCacheFetcher{logCacheClient: logCacheClient}
}

func (f LogCacheFetcher) FetchInstanceEntitlementUsages(appGuid string) ([]float64, error) {
	promqlResult, err := f.logCacheClient.PromQL(context.Background(), fmt.Sprintf(`absolute_usage{source_id="%s"} / absolute_entitlement{source_id="%s"}`, appGuid, appGuid))
	if err != nil {
		return nil, err
	}

	var instanceUsages []float64
	for _, sample := range promqlResult.GetVector().GetSamples() {
		instanceUsages = append(instanceUsages, sample.GetPoint().GetValue())
	}

	return instanceUsages, nil
}

func authenticatedBy(tokenGetter *token.Getter) *authClient {
	return &authClient{tokenGetter: tokenGetter}
}

type authClient struct {
	tokenGetter *token.Getter
}

func (a *authClient) Do(req *http.Request) (*http.Response, error) {
	t, err := a.tokenGetter.Token()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", t)
	return http.DefaultClient.Do(req)
}
