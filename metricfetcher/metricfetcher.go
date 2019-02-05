package metricfetcher // import "code.cloudfoundry.org/cpu-entitlement-plugin/metricfetcher"

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"code.cloudfoundry.org/cpu-entitlement-plugin/token"
	"code.cloudfoundry.org/cpu-entitlement-plugin/usagemetric"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	logcache "code.cloudfoundry.org/log-cache/pkg/client"
)

type CachedUsageMetricFetcher struct {
	client LogCacheClient
}

//go:generate counterfeiter . LogCacheClient
type LogCacheClient interface {
	Read(ctx context.Context, sourceID string, start time.Time, opts ...logcache.ReadOption) ([]*loggregator_v2.Envelope, error)
}

func New(logCacheURL string, tokenGetter *token.TokenGetter) CachedUsageMetricFetcher {
	return CachedUsageMetricFetcher{
		client: logcache.NewClient(
			logCacheURL,
			logcache.WithHTTPClient(authenticatedBy(tokenGetter)),
		),
	}
}

func NewWithLogCacheClient(client LogCacheClient) CachedUsageMetricFetcher {
	return CachedUsageMetricFetcher{client: client}
}

func (f CachedUsageMetricFetcher) FetchLatest(appGuid string, instanceCount int) ([]usagemetric.UsageMetric, error) {
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

func extractLatestMetrics(envelopes []*loggregator_v2.Envelope, instanceCount int) []usagemetric.UsageMetric {
	latestMetrics := make(map[int]usagemetric.UsageMetric, instanceCount)
	for _, envelope := range envelopes {
		usageMetric, ok := usagemetric.FromGaugeMetric(envelope.GetInstanceId(), envelope.GetGauge().GetMetrics())
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

func buildMetricsSlice(metricsMap map[int]usagemetric.UsageMetric, instanceCount int) []usagemetric.UsageMetric {
	var metrics []usagemetric.UsageMetric
	for i := 0; i < instanceCount; i++ {
		metric, ok := metricsMap[i]
		if ok {
			metrics = append(metrics, metric)
		}
	}

	return metrics
}

func authenticatedBy(tokenGetter *token.TokenGetter) *authClient {
	return &authClient{tokenGetter: tokenGetter}
}

type authClient struct {
	tokenGetter *token.TokenGetter
}

func (a *authClient) Do(req *http.Request) (*http.Response, error) {
	t, err := a.tokenGetter.Token()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", t)
	return http.DefaultClient.Do(req)
}
