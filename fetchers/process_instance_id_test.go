package fetchers_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cpu-entitlement-plugin/fetchers"
	"code.cloudfoundry.org/cpu-entitlement-plugin/fetchers/fetchersfakes"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Process Instance ID Fetcher", func() {

	var (
		fetcher            fetchers.ProcessInstanceIDFetcher
		logCacheClient     *fetchersfakes.FakeLogCacheClient
		processInstanceIDs map[int]string
		err                error
	)

	BeforeEach(func() {
		logCacheClient = new(fetchersfakes.FakeLogCacheClient)
		fetcher = fetchers.NewProcessInstanceIDFetcher(logCacheClient)
	})

	JustBeforeEach(func() {
		processInstanceIDs, err = fetcher.Fetch(logger, "the-app")
	})

	When("the first query returns #limit envelopes but doesn't return enough data to build the list of instances", func() {
		BeforeEach(func() {
			fetcher = fetchers.NewProcessInstanceIDFetcherWithLimit(logCacheClient, 3)
			logCacheClient.ReadReturns([]*loggregator_v2.Envelope{
				{
					InstanceId: "0",
					Tags: map[string]string{
						"process_instance_id": "instance-0-new",
					},
					Timestamp: 10,
				},
				{
					InstanceId: "2",
					Tags: map[string]string{
						"process_instance_id": "instance-2",
					},
					Timestamp: 9,
				},
				{
					InstanceId: "0",
					Tags: map[string]string{
						"process_instance_id": "instance-0-old",
					},
					Timestamp: time.Now().Add(-15 * time.Second).UnixNano(),
				},
			}, nil)

		})

		It("will exit having called Read 10 times, which is our sanity limit", func() {
			Expect(logCacheClient.ReadCallCount()).To(Equal(10))
		})

		Context("When read#9 returns fewer than limit", func() {

			BeforeEach(func() {
				logCacheClient.ReadReturnsOnCall(8, []*loggregator_v2.Envelope{
					{
						InstanceId: "1",
						Tags: map[string]string{
							"process_instance_id": "instance-1",
						},
						Timestamp: 7,
					},
					{
						InstanceId: "2",
						Tags: map[string]string{
							"process_instance_id": "instance-2-old",
						},
						Timestamp: 6,
					},
				}, nil)
			})

			It("calls Read() 9 times then stops", func() {
				Expect(logCacheClient.ReadCallCount()).To(Equal(9))
				Expect(processInstanceIDs).To(Equal(map[int]string{
					0: "instance-0-new",
					1: "instance-1",
					2: "instance-2",
				}))
			})
		})
	})

	When("reading from logcache fails", func() {
		BeforeEach(func() {
			logCacheClient.ReadReturns(nil, errors.New("logcache-error"))
		})

		It("fails", func() {
			Expect(err).To(MatchError("logcache-error"))
		})
	})

	When("the instance id is not an int", func() {
		BeforeEach(func() {
			logCacheClient.ReadReturns([]*loggregator_v2.Envelope{
				{
					InstanceId: "0f",
					Tags: map[string]string{
						"process_instance_id": "instance-0-borked",
					},
					Timestamp: 15,
				},
				{
					InstanceId: "0",
					Tags: map[string]string{
						"process_instance_id": "instance-0",
					},
					Timestamp: 10,
				},
			}, nil)
		})

		It("should ignore the metric with the invalid instance id", func() {
			Expect(processInstanceIDs).To(Equal(map[int]string{0: "instance-0"}))
		})
	})

	When("the process instance id is empty or not set", func() {
		BeforeEach(func() {
			logCacheClient.ReadReturns([]*loggregator_v2.Envelope{
				{
					InstanceId: "0",
					Tags:       map[string]string{},
					Timestamp:  15,
				},
				{
					InstanceId: "0",
					Tags:       map[string]string{"process_instance_id": ""},
					Timestamp:  13,
				},
				{
					InstanceId: "0",
					Tags: map[string]string{
						"process_instance_id": "instance-0",
					},
					Timestamp: 10,
				},
			}, nil)
		})

		It("should ignore the metric with the missing process instance id", func() {
			Expect(processInstanceIDs).To(Equal(map[int]string{0: "instance-0"}))
		})
	})
})
