package logstreamer

import (
	"context"
	"log"
	"net/http"
	"os"

	loggregator "code.cloudfoundry.org/go-loggregator"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/cloudfoundry/cpu-entitlement-plugin/usagemetric"
)

type LogStreamer struct {
	client LoggregatorClient
}

//go:generate counterfeiter . LoggregatorClient
type LoggregatorClient interface {
	Stream(ctx context.Context, req *loggregator_v2.EgressBatchRequest) loggregator.EnvelopeStream
}

func New(logStreamURL, token string) LogStreamer {
	return LogStreamer{
		client: loggregator.NewRLPGatewayClient(
			logStreamURL,
			loggregator.WithRLPGatewayClientLogger(log.New(os.Stderr, "", log.LstdFlags)),
			loggregator.WithRLPGatewayHTTPClient(authenticatedBy(token)),
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
			usageMetric, ok := usagemetric.FromGaugeMetric(envelope.GetGauge().GetMetrics())
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

func authenticatedBy(token string) *authClient {
	return &authClient{token: token}
}

type authClient struct {
	token string
}

func (a *authClient) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", a.token)
	return http.DefaultClient.Do(req)
}
