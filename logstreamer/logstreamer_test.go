package logstreamer_test

import (
	"context"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cpu-entitlement-plugin/logstreamer"
	"github.com/cloudfoundry/cpu-entitlement-plugin/logstreamer/logstreamerfakes"
	"github.com/cloudfoundry/cpu-entitlement-plugin/usagemetric"
)

var _ = Describe("Logstreamer", func() {
	var (
		loggregatorClient *logstreamerfakes.FakeLoggregatorClient
		logStreamer       LogStreamer
		appGuid           string
		metrics           chan usagemetric.UsageMetric
	)

	BeforeEach(func() {
		loggregatorClient = new(logstreamerfakes.FakeLoggregatorClient)
		logStreamer = NewWithLoggregatorClient(loggregatorClient)

		appGuid = "foo"

		currentBatchIndex := 0
		batches := [][]*loggregator_v2.Envelope{
			{
				MetricEnvelope(appGuid, Metric{Usage: 1000, Entitlement: 5000, Age: 10000}),
				MetricEnvelope(appGuid, Metric{Usage: 2000, Entitlement: 6000, Age: 11000}),
				MetricEnvelope(appGuid, Metric{Usage: 3000, Entitlement: 7000, Age: 12000}),
			},
			{
				OtherEnvelope(appGuid),
				MetricEnvelope(appGuid, Metric{Usage: 4000, Entitlement: 8000, Age: 13000}),
				MetricEnvelope(appGuid, Metric{Usage: 5000, Entitlement: 9000, Age: 14000}),
			},
		}

		streamFunc := func() []*loggregator_v2.Envelope {
			if currentBatchIndex < len(batches) {
				currentBatch := batches[currentBatchIndex]
				currentBatchIndex += 1
				return currentBatch
			}
			return nil
		}

		loggregatorClient.StreamReturns(streamFunc)
	})

	JustBeforeEach(func() {
		metrics = logStreamer.Stream(appGuid)
	})

	It("gets the stream from the loggregator client", func() {
		Expect(loggregatorClient.StreamCallCount()).To(Equal(1))
		ctx, streamReq := loggregatorClient.StreamArgsForCall(0)
		Expect(ctx).To(Equal(context.Background()))
		Expect(streamReq).To(Equal(&loggregator_v2.EgressBatchRequest{
			Selectors: []*loggregator_v2.Selector{
				{
					SourceId: appGuid,
					Message:  &loggregator_v2.Selector_Gauge{},
				},
			}}))
	})

	It("returns the correct metrics", func() {
		receivedMetrics := []usagemetric.UsageMetric{}
		for i := 0; i < 5; i++ {
			receivedMetrics = append(receivedMetrics, <-metrics)
		}

		Consistently(metrics).ShouldNot(Receive())

		Expect(receivedMetrics).To(ConsistOf(
			usagemetric.UsageMetric{AbsoluteUsage: 1000, AbsoluteEntitlement: 5000, ContainerAge: 10000},
			usagemetric.UsageMetric{AbsoluteUsage: 2000, AbsoluteEntitlement: 6000, ContainerAge: 11000},
			usagemetric.UsageMetric{AbsoluteUsage: 3000, AbsoluteEntitlement: 7000, ContainerAge: 12000},
			usagemetric.UsageMetric{AbsoluteUsage: 4000, AbsoluteEntitlement: 8000, ContainerAge: 13000},
			usagemetric.UsageMetric{AbsoluteUsage: 5000, AbsoluteEntitlement: 9000, ContainerAge: 14000},
		))
	})
})

type Metric struct {
	Usage       float64
	Entitlement float64
	Age         float64
}

func MetricEnvelope(appGuid string, metric Metric) *loggregator_v2.Envelope {
	return &loggregator_v2.Envelope{
		SourceId: appGuid,
		Message: &loggregator_v2.Envelope_Gauge{
			Gauge: &loggregator_v2.Gauge{
				Metrics: map[string]*loggregator_v2.GaugeValue{
					"absolute_usage":       &loggregator_v2.GaugeValue{Value: metric.Usage},
					"absolute_entitlement": &loggregator_v2.GaugeValue{Value: metric.Entitlement},
					"container_age":        &loggregator_v2.GaugeValue{Value: metric.Age},
				},
			},
		},
	}
}

func OtherEnvelope(appGuid string) *loggregator_v2.Envelope {
	return &loggregator_v2.Envelope{
		SourceId: appGuid,
		Message: &loggregator_v2.Envelope_Gauge{
			Gauge: &loggregator_v2.Gauge{
				Metrics: map[string]*loggregator_v2.GaugeValue{
					"foo": &loggregator_v2.GaugeValue{Value: 42},
				},
			},
		},
	}
}
