package logstreamer // import "code.cloudfoundry.org/cpu-entitlement-plugin/logstreamer"

import (
	"context"
	"log"
	"net/http"
	"os"

	"code.cloudfoundry.org/cpu-entitlement-plugin/token"
	"code.cloudfoundry.org/cpu-entitlement-plugin/usagemetric"
	loggregator "code.cloudfoundry.org/go-loggregator"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
)

type LogStreamer struct {
	client LoggregatorClient
}

//go:generate counterfeiter . LoggregatorClient
type LoggregatorClient interface {
	Stream(ctx context.Context, req *loggregator_v2.EgressBatchRequest) loggregator.EnvelopeStream
}

func New(logStreamURL string, tokenGetter *token.TokenGetter) LogStreamer {
	return LogStreamer{
		client: loggregator.NewRLPGatewayClient(
			logStreamURL,
			loggregator.WithRLPGatewayClientLogger(log.New(os.Stderr, "", log.LstdFlags)),
			loggregator.WithRLPGatewayHTTPClient(authenticatedBy(tokenGetter)),
		),
	}
}

func NewWithLoggregatorClient(client LoggregatorClient) LogStreamer {
	return LogStreamer{client: client}
}

func (s LogStreamer) Stream(appGuid string) chan usagemetric.UsageMetric {
	stream := s.client.Stream(context.Background(), streamRequest(appGuid))

	var usageMetricsStream = make(chan usagemetric.UsageMetric)
	go streamToUsageChan(stream, usageMetricsStream)

	return usageMetricsStream
}

func streamToUsageChan(stream loggregator.EnvelopeStream, usageMetricsStream chan<- usagemetric.UsageMetric) {
	for {
		for _, envelope := range stream() {
			usageMetric, ok := usagemetric.FromGaugeMetric(envelope.GetInstanceId(), envelope.GetGauge().GetMetrics())
			if !ok {
				continue
			}

			usageMetricsStream <- usageMetric
		}
	}
}

func streamRequest(sourceID string) *loggregator_v2.EgressBatchRequest {
	return &loggregator_v2.EgressBatchRequest{
		Selectors: []*loggregator_v2.Selector{
			{
				SourceId: sourceID,
				Message:  &loggregator_v2.Selector_Gauge{},
			},
		},
	}
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
