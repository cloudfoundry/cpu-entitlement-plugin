package metrics_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"code.cloudfoundry.org/log-cache/pkg/rpc/logcache_v1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cpu-entitlement-plugin/metrics"
	"code.cloudfoundry.org/cpu-entitlement-plugin/metrics/metricsfakes"
)

var _ = Describe("Logstreamer", func() {
	var (
		logCacheClient     *metricsfakes.FakeLogCacheClient
		metricsFetcher     metrics.LogCacheFetcher
		appGuid            string
		instanceDataPoints map[int][]metrics.InstanceData
		metricsErr         error
		from, to           time.Time
	)

	BeforeEach(func() {
		logCacheClient = new(metricsfakes.FakeLogCacheClient)
		metricsFetcher = metrics.NewFetcherWithLogCacheClient(logCacheClient)

		appGuid = "foo"
		from = time.Now().Add(-time.Hour)
		to = time.Now()

		logCacheClient.PromQLReturns(instantQueryResult(
			sample("abc"),
			sample("def"),
			sample("ghi"),
		), nil)
	})

	JustBeforeEach(func() {
		instanceDataPoints, metricsErr = metricsFetcher.FetchInstanceData(appGuid, from, to)
	})

	When("reading from log-cache succeeds", func() {
		BeforeEach(func() {
			logCacheClient.PromQLRangeReturns(rangeQueryResult(
				series("0", "abc",
					point("1", 0.2),
					point("3", 0.5),
				),
				series("1", "def",
					point("2", 0.4),
				),
				series("2", "ghi",
					point("4", 0.5),
				),
			), nil)
		})

		It("gets the metrics from the log-cache client", func() {
			Expect(logCacheClient.PromQLCallCount()).To(Equal(1))
			ctx, query, _ := logCacheClient.PromQLArgsForCall(0)
			Expect(ctx).To(Equal(context.Background()))
			Expect(query).To(Equal(fmt.Sprintf(`absolute_usage{source_id="%s"}`, appGuid)))

			Expect(logCacheClient.PromQLRangeCallCount()).To(Equal(1))
			ctx, query, _ = logCacheClient.PromQLRangeArgsForCall(0)
			Expect(ctx).To(Equal(context.Background()))
			Expect(query).To(Equal(fmt.Sprintf(`absolute_usage{source_id="%s"} / absolute_entitlement{source_id="%s"}`, appGuid, appGuid)))
		})

		It("returns the correct metrics", func() {
			Expect(metricsErr).NotTo(HaveOccurred())
			Expect(instanceDataPoints).To(Equal(map[int][]metrics.InstanceData{
				0: {
					{
						Time:             time.Unix(1, 0),
						InstanceID:       0,
						EntitlementUsage: 0.2,
					},
					{
						Time:             time.Unix(3, 0),
						InstanceID:       0,
						EntitlementUsage: 0.5,
					},
				},
				1: {
					{
						Time:             time.Unix(2, 0),
						InstanceID:       1,
						EntitlementUsage: 0.4,
					},
				},
				2: {
					{
						Time:             time.Unix(4, 0),
						InstanceID:       2,
						EntitlementUsage: 0.5,
					},
				},
			}))
		})
	})

	When("fetching the list of active instances from log-cache fails", func() {
		BeforeEach(func() {
			logCacheClient.PromQLReturns(nil, errors.New("boo"))
		})

		It("returns an error", func() {
			Expect(metricsErr).To(MatchError("boo"))
			Expect(instanceDataPoints).To(BeNil())
		})
	})

	When("fetching the list of data points from log-cache fails", func() {
		BeforeEach(func() {
			logCacheClient.PromQLRangeReturns(nil, errors.New("boo"))
		})

		It("returns an error", func() {
			Expect(metricsErr).To(MatchError("boo"))
			Expect(instanceDataPoints).To(BeNil())
		})
	})

	When("getting stale data from old instances after a scale down", func() {
		BeforeEach(func() {
			logCacheClient.PromQLRangeReturns(rangeQueryResult(
				series("0", "abc",
					point("1", 0.5),
				),
				series("1", "def",
					point("2", 0.5),
				),
				series("2", "ghi",
					point("3", 0.5),
				),
				series("3", "jkl",
					point("4", 0.5),
				),
			), nil)
		})

		It("ignores the data from old instances", func() {
			Expect(instanceDataPoints).To(Equal(map[int][]metrics.InstanceData{
				0: {
					{
						Time:             time.Unix(1, 0),
						InstanceID:       0,
						EntitlementUsage: 0.5,
					},
				},
				1: {
					{
						Time:             time.Unix(2, 0),
						InstanceID:       1,
						EntitlementUsage: 0.5,
					},
				},
				2: {
					{
						Time:             time.Unix(3, 0),
						InstanceID:       2,
						EntitlementUsage: 0.5,
					},
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

func instantQueryResult(samples ...*logcache_v1.PromQL_Sample) *logcache_v1.PromQL_InstantQueryResult {
	return &logcache_v1.PromQL_InstantQueryResult{
		Result: &logcache_v1.PromQL_InstantQueryResult_Vector{
			Vector: &logcache_v1.PromQL_Vector{
				Samples: samples,
			},
		},
	}
}

func sample(procInstanceID string) *logcache_v1.PromQL_Sample {
	return &logcache_v1.PromQL_Sample{
		Metric: map[string]string{
			"process_instance_id": procInstanceID,
		},
	}
}

func rangeQueryResult(series ...*logcache_v1.PromQL_Series) *logcache_v1.PromQL_RangeQueryResult {
	return &logcache_v1.PromQL_RangeQueryResult{
		Result: &logcache_v1.PromQL_RangeQueryResult_Matrix{
			Matrix: &logcache_v1.PromQL_Matrix{
				Series: series,
			},
		},
	}
}

func series(instanceID, procInstanceID string, points ...*logcache_v1.PromQL_Point) *logcache_v1.PromQL_Series {
	return &logcache_v1.PromQL_Series{
		Metric: map[string]string{
			"instance_id":         instanceID,
			"process_instance_id": procInstanceID,
		},
		Points: points,
	}
}

func point(time string, value float64) *logcache_v1.PromQL_Point {
	return &logcache_v1.PromQL_Point{Time: time, Value: value}
}
