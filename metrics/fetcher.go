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

func (f LogCacheFetcher) FetchLatest(appGuid string, instanceCount int) ([]Usage, error) {
	envelopes, err := f.client.Read(context.Background(), appGuid, time.Now().Add(-5*time.Minute), logcache.WithDescending())
	if err != nil {
		return nil, fmt.Errorf("log-cache read failed: %s", err.Error())
	}

	latestMetrics := extractLatestMetrics(envelopes, instanceCount)
	if len(latestMetrics) == 0 {
		return nil, fmt.Errorf("No CPU metrics found for '%s'", appGuid)
	}

	return latestMetrics, nil
}

func extractLatestMetrics(envelopes []*loggregator_v2.Envelope, instanceCount int) []Usage {
	latestMetrics := make(map[int]Usage, instanceCount)
	for _, envelope := range envelopes {
		usageMetric, ok := UsageFromGauge(envelope.GetInstanceId(), envelope.GetGauge().GetMetrics())
		if ok && usageMetric.InstanceId < instanceCount {
			_, exists := latestMetrics[usageMetric.InstanceId]
			if !exists {
				latestMetrics[usageMetric.InstanceId] = usageMetric
			}

			if len(latestMetrics) == instanceCount {
				break
			}
		}
	}

	return buildMetricsSlice(latestMetrics, instanceCount)
}

func buildMetricsSlice(metricsMap map[int]Usage, instanceCount int) []Usage {
	var metrics []Usage
	for i := 0; i < instanceCount; i++ {
		metric, ok := metricsMap[i]
		if ok {
			metrics = append(metrics, metric)
		}
	}

	return metrics
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
