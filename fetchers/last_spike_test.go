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
		spikes         map[int][]fetchers.InstanceData
		fetchErr       error
		since          time.Time
	)

	BeforeEach(func() {
		appGuid = "foo"
		since = time.Now().Add(-time.Hour)
		logCacheClient = new(fetchersfakes.FakeLogCacheClient)
		fetcher = *fetchers.NewLastSpikeFetcher(logCacheClient, since)

		logCacheClient.ReadReturns([]*loggregator_v2.Envelope{
			{
				InstanceId: "0",
				Tags: map[string]string{
					"process_instance_id": "abc",
				},
				Timestamp: 13,
				Message: &loggregator_v2.Envelope_Gauge{
					Gauge: &loggregator_v2.Gauge{
						Metrics: map[string]*loggregator_v2.GaugeValue{
							"spike_start": &loggregator_v2.GaugeValue{Value: 3},
							"spike_end":   &loggregator_v2.GaugeValue{Value: 4},
						},
					},
				},
			},
			{
				InstanceId: "1",
				Tags: map[string]string{
					"process_instance_id": "def",
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

		appInstances = map[int]cf.Instance{
			0: cf.Instance{InstanceID: 0, ProcessInstanceID: "abc"},
			1: cf.Instance{InstanceID: 1, ProcessInstanceID: "def"},
		}
	})

	JustBeforeEach(func() {
		spikes, fetchErr = fetcher.FetchInstanceData(appGuid, appInstances)
	})

	It("gets the last spike from the log-cache client", func() {
		Expect(logCacheClient.ReadCallCount()).To(Equal(1))
		_, actualSourceID, actualTime, _ := logCacheClient.ReadArgsForCall(0)
		Expect(actualSourceID).To(Equal(appGuid))
		Expect(actualTime).To(Equal(since))
	})

	It("returns the last spike", func() {
		Expect(fetchErr).NotTo(HaveOccurred())

		Expect(spikes).To(HaveKey(0))
		Expect(spikes[0]).To(HaveLen(1))
		Expect(spikes[0][0]).To(Equal(fetchers.InstanceData{
			InstanceID: 0,
			From:       time.Unix(3, 0),
			To:         time.Unix(4, 0),
		}))

		Expect(spikes).To(HaveKey(1))
		Expect(spikes[1]).To(HaveLen(1))
		Expect(spikes[1][0]).To(Equal(fetchers.InstanceData{
			InstanceID: 1,
			From:       time.Unix(5, 0),
			To:         time.Unix(6, 0),
		}))
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
					InstanceId: "1",
					Tags: map[string]string{
						"process_instance_id": "def",
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
			Expect(spikes).To(HaveKey(1))
		})
	})

	When("the envelope's message is not a gauge", func() {
		BeforeEach(func() {
			logCacheClient.ReadReturns([]*loggregator_v2.Envelope{
				{
					InstanceId: "0",
					Tags: map[string]string{
						"process_instance_id": "abc",
					},
					Timestamp: 13,
					Message:   &loggregator_v2.Envelope_Timer{},
				},
				{
					InstanceId: "1",
					Tags: map[string]string{
						"process_instance_id": "def",
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
			Expect(spikes).To(HaveKey(1))
		})
	})

	When("cache returns data for instances that are no longer running (because the app has been scaled down", func() {
		BeforeEach(func() {
			appInstances = map[int]cf.Instance{
				0: cf.Instance{InstanceID: 0, ProcessInstanceID: "abc"},
			}
		})

		It("returns historical usage for running instances only", func() {
			Expect(fetchErr).NotTo(HaveOccurred())

			Expect(spikes).To(HaveKey(0))
			Expect(spikes[0]).To(HaveLen(1))
			Expect(spikes[0][0]).To(Equal(fetchers.InstanceData{
				InstanceID: 0,
				From:       time.Unix(3, 0),
				To:         time.Unix(4, 0),
			}))

			Expect(spikes).NotTo(HaveKey(1))
		})
	})

	When("cache returns data for instances with same id but different process instance id", func() {
		BeforeEach(func() {
			appInstances = map[int]cf.Instance{
				0: cf.Instance{InstanceID: 0, ProcessInstanceID: "not-abc"},
			}
		})

		It("ignores that data", func() {
			Expect(spikes).To(BeEmpty())
		})
	})
})

func MetricEnvelope(appGuid, instanceId string, metric Metric) *loggregator_v2.Envelope {
	return &loggregator_v2.Envelope{
		SourceId:   appGuid,
		InstanceId: instanceId,
	}
}
