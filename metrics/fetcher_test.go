package metrics_test

import (
	"context"
	"errors"
	"time"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cpu-entitlement-plugin/metrics"
	"code.cloudfoundry.org/cpu-entitlement-plugin/metrics/metricsfakes"
)

var _ = Describe("Logstreamer", func() {
	var (
		logCacheClient *metricsfakes.FakeLogCacheClient
		metricsFetcher metrics.LogCacheFetcher
		appGuid        string
		usageMetrics   []metrics.InstanceData
		metricsErr     error
	)

	BeforeEach(func() {
		logCacheClient = new(metricsfakes.FakeLogCacheClient)
		metricsFetcher = metrics.NewFetcherWithLogCacheClient(logCacheClient)

		appGuid = "foo"
	})

	JustBeforeEach(func() {
		usageMetrics, metricsErr = metricsFetcher.FetchAll(appGuid, 3)
	})

	When("reading from log-cache succeeds", func() {
		BeforeEach(func() {
			logCacheClient.ReadReturns([]*loggregator_v2.Envelope{
				OtherEnvelope(appGuid),
				MetricEnvelope(1, appGuid, "0", Metric{Usage: 1000, Entitlement: 5000, Age: 10000}),
				MetricEnvelope(2, appGuid, "1", Metric{Usage: 2000, Entitlement: 6000, Age: 11000}),
				MetricEnvelope(3, appGuid, "0", Metric{Usage: 3000, Entitlement: 7000, Age: 12000}),
				MetricEnvelope(4, appGuid, "2", Metric{Usage: 4000, Entitlement: 8000, Age: 13000}),
			}, nil)
		})

		It("gets the metrics from the log-cache client", func() {
			Expect(logCacheClient.ReadCallCount()).To(Equal(1))
			ctx, sourceID, startTime, _ := logCacheClient.ReadArgsForCall(0)
			Expect(ctx).To(Equal(context.Background()))
			Expect(sourceID).To(Equal(appGuid))
			Expect(startTime).To(BeTemporally("~", time.Now().Add(-metrics.Month)))
		})

		It("returns the correct metrics", func() {
			Expect(metricsErr).NotTo(HaveOccurred())
			Expect(usageMetrics).To(Equal([]metrics.InstanceData{
				{
					Time:                time.Unix(0, 1),
					InstanceId:          0,
					AbsoluteUsage:       1000,
					AbsoluteEntitlement: 5000,
					ContainerAge:        10000,
				},
				{
					Time:                time.Unix(0, 2),
					InstanceId:          1,
					AbsoluteUsage:       2000,
					AbsoluteEntitlement: 6000,
					ContainerAge:        11000,
				},
				{
					Time:                time.Unix(0, 3),
					InstanceId:          0,
					AbsoluteUsage:       3000,
					AbsoluteEntitlement: 7000,
					ContainerAge:        12000,
				},
				{
					Time:                time.Unix(0, 4),
					InstanceId:          2,
					AbsoluteUsage:       4000,
					AbsoluteEntitlement: 8000,
					ContainerAge:        13000,
				},
			}))
		})
	})

	When("reading from log-cache fails", func() {
		BeforeEach(func() {
			logCacheClient.ReadReturns(nil, errors.New("boo"))
		})

		It("returns an error", func() {
			Expect(metricsErr).To(MatchError("log-cache read failed: boo"))
			Expect(usageMetrics).To(BeNil())
		})
	})

	When("no usage metrics envelopes are returned", func() {
		BeforeEach(func() {
			logCacheClient.ReadReturns([]*loggregator_v2.Envelope{
				OtherEnvelope(appGuid),
				OtherEnvelope(appGuid),
				OtherEnvelope(appGuid),
			}, nil)
		})

		It("returns an error", func() {
			Expect(metricsErr).To(MatchError("No CPU metrics found for '" + appGuid + "'"))
			Expect(usageMetrics).To(BeNil())
		})
	})

	When("getting stale data from old instances after a scale down", func() {
		BeforeEach(func() {
			logCacheClient.ReadReturns([]*loggregator_v2.Envelope{
				MetricEnvelope(1, appGuid, "3", Metric{Usage: 5000, Entitlement: 9000, Age: 14000}),
				MetricEnvelope(2, appGuid, "2", Metric{Usage: 4000, Entitlement: 8000, Age: 13000}),
				MetricEnvelope(3, appGuid, "1", Metric{Usage: 2000, Entitlement: 6000, Age: 11000}),
				MetricEnvelope(4, appGuid, "0", Metric{Usage: 1000, Entitlement: 5000, Age: 10000}),
			}, nil)
		})

		It("ignores the data from old instances", func() {
			Expect(usageMetrics).To(Equal([]metrics.InstanceData{
				{
					Time:                time.Unix(0, 2),
					InstanceId:          2,
					AbsoluteUsage:       4000,
					AbsoluteEntitlement: 8000,
					ContainerAge:        13000,
				},
				{
					Time:                time.Unix(0, 3),
					InstanceId:          1,
					AbsoluteUsage:       2000,
					AbsoluteEntitlement: 6000,
					ContainerAge:        11000,
				},
				{
					Time:                time.Unix(0, 4),
					InstanceId:          0,
					AbsoluteUsage:       1000,
					AbsoluteEntitlement: 5000,
					ContainerAge:        10000,
				},
			}))
		})
	})
})

type Metric struct {
	Usage       float64
	Entitlement float64
	Age         float64
}

func MetricEnvelope(timestamp int64, appGuid, instanceId string, metric Metric) *loggregator_v2.Envelope {
	return &loggregator_v2.Envelope{
		Timestamp:  timestamp,
		SourceId:   appGuid,
		InstanceId: instanceId,
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
