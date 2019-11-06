package fetchers_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cpu-entitlement-plugin/cf"
	"code.cloudfoundry.org/cpu-entitlement-plugin/fetchers"
	"code.cloudfoundry.org/cpu-entitlement-plugin/fetchers/fetchersfakes"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
)

var _ = Describe("LastSpikeFetcher", func() {
	var (
		logCacheClient *fetchersfakes.FakeLogCacheClient
		fetcher        fetchers.LastSpikeFetcher
		appGuid        string
		appInstances   map[int]cf.Instance
		spikes         map[int]interface{}
		fetchErr       error
		since          time.Time
	)

	BeforeEach(func() {
		appGuid = "foo"
		since = time.Now().Add(-time.Hour)
		logCacheClient = new(fetchersfakes.FakeLogCacheClient)
		fetcher = *fetchers.NewLastSpikeFetcher(logCacheClient, since)

		appInstances = map[int]cf.Instance{
			0: cf.Instance{InstanceID: 0, ProcessInstanceID: "abc"},
		}
	})

	JustBeforeEach(func() {
		spikes, fetchErr = fetcher.FetchInstanceData(appGuid, appInstances)
	})

	When("fetching the list of data points from log-cache fails", func() {
		BeforeEach(func() {
			logCacheClient.ReadReturns(nil, errors.New("boo"))
		})

		It("returns an error", func() {
			Expect(fetchErr).To(MatchError("boo"))
			Expect(spikes).To(BeNil())
		})
	})

	When("the instance ID is not a valid number", func() {
		BeforeEach(func() {
			logCacheClient.ReadReturns([]*loggregator_v2.Envelope{
				{
					InstanceId: "not-valid",
				},
				{
					InstanceId: "0",
					Tags: map[string]string{
						"process_instance_id": "abc",
					},
					Timestamp: 10,
					Message: &loggregator_v2.Envelope_Gauge{
						Gauge: &loggregator_v2.Gauge{
							Metrics: map[string]*loggregator_v2.GaugeValue{
								"spike_start": &loggregator_v2.GaugeValue{Value: 5},
								"spike_end":   &loggregator_v2.GaugeValue{Value: 6},
							},
						},
					},
				},
			}, nil)
		})

		It("ignores the invalid entries", func() {
			Expect(fetchErr).NotTo(HaveOccurred())
			Expect(spikes).To(HaveLen(1))
			Expect(spikes).To(HaveKey(0))
		})
	})
})

func MetricEnvelope(appGuid, instanceId string, metric Metric) *loggregator_v2.Envelope {
	return &loggregator_v2.Envelope{
		SourceId:   appGuid,
		InstanceId: instanceId,
	}
}
