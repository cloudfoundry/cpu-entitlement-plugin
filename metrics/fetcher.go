package metrics // import "code.cloudfoundry.org/cpu-entitlement-plugin/metrics"

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"code.cloudfoundry.org/cpu-entitlement-plugin/token"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	logcache "code.cloudfoundry.org/log-cache/pkg/client"
)

const Month = 730 * time.Hour

type LogCacheFetcher struct {
	client LogCacheClient
}

//go:generate counterfeiter . LogCacheClient
type LogCacheClient interface {
	Read(ctx context.Context, sourceID string, start time.Time, opts ...logcache.ReadOption) ([]*loggregator_v2.Envelope, error)
}

func NewFetcher(logCacheURL string, tokenGetter *token.Getter) LogCacheFetcher {
	return LogCacheFetcher{
		client: logcache.NewClient(
			logCacheURL,
			logcache.WithHTTPClient(authenticatedBy(tokenGetter)),
		),
	}
}

func NewFetcherWithLogCacheClient(client LogCacheClient) LogCacheFetcher {
	return LogCacheFetcher{client: client}
}

func (f LogCacheFetcher) FetchAll(appGuid string, instanceCount int) ([]InstanceData, error) {
	envelopes, err := f.client.Read(context.Background(), appGuid, time.Now().Add(-Month), logcache.WithDescending())
	if err != nil {
		return nil, fmt.Errorf("log-cache read failed: %s", err.Error())
	}

	instancesData := parseEnvelopes(envelopes, instanceCount)
	if len(instancesData) == 0 {
		return nil, fmt.Errorf("No CPU metrics found for '%s'", appGuid)
	}

	return instancesData, nil
}

func parseEnvelopes(envelopes []*loggregator_v2.Envelope, instanceCount int) []InstanceData {
	var instancesData []InstanceData
	for _, envelope := range envelopes {
		instanceData, ok := InstanceDataFromGauge(envelope.GetInstanceId(), envelope.GetGauge().GetMetrics())
		if ok && instanceData.InstanceId < instanceCount {
			instancesData = append(instancesData, instanceData)
		}
	}

	return instancesData
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
