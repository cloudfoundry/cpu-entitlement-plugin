package fetchers_test

import (
	"context"
	"errors"
	"fmt"

	"code.cloudfoundry.org/cpu-entitlement-plugin/cf"
	"code.cloudfoundry.org/cpu-entitlement-plugin/fetchers"
	"code.cloudfoundry.org/cpu-entitlement-plugin/fetchers/fetchersfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CurrentUsage", func() {
	var (
		logCacheClient    *fetchersfakes.FakeLogCacheClient
		historicalFetcher *fetchersfakes.FakeHistoricalFetcher
		fetcher           fetchers.CurrentUsageFetcher
		appGuid           string
		appInstances      map[int]cf.Instance
		currentUsage      map[int][]fetchers.InstanceData
		fetchErr          error
	)

	BeforeEach(func() {
		logCacheClient = new(fetchersfakes.FakeLogCacheClient)
		historicalFetcher = new(fetchersfakes.FakeHistoricalFetcher)
		fetcher = fetchers.NewCurrentUsageFetcherWithHistoricalFetcher(logCacheClient, historicalFetcher)

		appGuid = "foo"

		logCacheClient.PromQLReturns(queryResult(
			sample("0", "abc",
				point("1", 0.2),
			),
			sample("1", "def",
				point("2", 0.4),
			),
			sample("2", "ghi",
				point("4", 0.5),
			),
		), nil)
		appInstances = map[int]cf.Instance{
			0: cf.Instance{InstanceID: 0, ProcessInstanceID: "abc"},
			1: cf.Instance{InstanceID: 1, ProcessInstanceID: "def"},
			2: cf.Instance{InstanceID: 2, ProcessInstanceID: "ghi"},
		}
	})

	JustBeforeEach(func() {
		currentUsage, fetchErr = fetcher.FetchInstanceData(appGuid, appInstances)
	})

	It("gets the current usage from the log-cache client", func() {
		Expect(logCacheClient.PromQLCallCount()).To(Equal(1))
		ctx, query, _ := logCacheClient.PromQLArgsForCall(0)
		Expect(ctx).To(Equal(context.Background()))
		Expect(query).To(Equal(fmt.Sprintf(`idelta(absolute_usage{source_id="%s"}[1m]) / idelta(absolute_entitlement{source_id="%s"}[1m])`, appGuid, appGuid)))
	})

	When("result does not contain points for every instance", func() {
		BeforeEach(func() {
			logCacheClient.PromQLReturns(queryResult(
				sample("0", "abc",
					point("1", 0.2),
				),
				sample("1", "def",
					point("2", 0.4),
				),
			), nil)
			historicalFetcher.FetchInstanceDataReturns(map[int][]fetchers.InstanceData{
				2: {
					fetchers.InstanceData{InstanceID: 2, Value: 1.1},
					fetchers.InstanceData{InstanceID: 2, Value: 1.4},
				},
			}, nil)
		})

		It("fetches historical data", func() {
			Expect(historicalFetcher.FetchInstanceDataCallCount()).To(Equal(1))
			actualAppGuid, actualAppInstances := historicalFetcher.FetchInstanceDataArgsForCall(0)
			Expect(actualAppGuid).To(Equal(appGuid))
			Expect(actualAppInstances).To(Equal(appInstances))
		})

		It("succeeds", func() {
			Expect(fetchErr).NotTo(HaveOccurred())
		})

		It("returns values for each of the app instances", func() {
			Expect(len(currentUsage)).To(Equal(3))
			Expect(currentUsage).To(HaveKey(0))
			Expect(currentUsage).To(HaveKey(1))
			Expect(currentUsage).To(HaveKey(2))
		})

		It("returns the values from the delta query for the instances that have such", func() {
			Expect(currentUsage[0]).To(ConsistOf(fetchers.InstanceData{
				InstanceID: 0,
				Value:      0.2,
			}))

			Expect(currentUsage[1]).To(ConsistOf(fetchers.InstanceData{
				InstanceID: 1,
				Value:      0.4,
			}))
		})

		It("returns the values from the range query for the instances that don't have a delta value", func() {
			Expect(currentUsage[2]).To(Equal([]fetchers.InstanceData{
				{
					InstanceID: 2,
					Value:      1.4,
				},
			}))
		})

		When("fetching historical data fails", func() {
			BeforeEach(func() {
				historicalFetcher.FetchInstanceDataReturns(nil, errors.New("fetch-failed"))
			})

			It("returns the error", func() {
				Expect(fetchErr).To(MatchError("fetch-failed"))
			})
		})

	})

	It("returns the correct current usage", func() {
		Expect(fetchErr).NotTo(HaveOccurred())
		Expect(currentUsage).To(Equal(map[int][]fetchers.InstanceData{
			0: {
				{
					InstanceID: 0,
					Value:      0.2,
				},
			},
			1: {
				{
					InstanceID: 1,
					Value:      0.4,
				},
			},
			2: {
				{
					InstanceID: 2,
					Value:      0.5,
				},
			},
		}))
	})

	When("cache returns data for instances that are no longer running (because the app has been scaled down", func() {
		BeforeEach(func() {
			appInstances = map[int]cf.Instance{
				0: cf.Instance{InstanceID: 0, ProcessInstanceID: "abc"},
			}
		})

		It("returns current usage for running instances only", func() {
			Expect(fetchErr).NotTo(HaveOccurred())
			Expect(currentUsage).To(Equal(map[int][]fetchers.InstanceData{
				0: {
					{
						InstanceID: 0,
						Value:      0.2,
					},
				},
			}))
		})
	})

	When("cache returns data for instances with same id but different process instance id", func() {
		BeforeEach(func() {
			logCacheClient.PromQLReturns(queryResult(
				sample("0", "def",
					point("1", 0.2),
				),
			), nil)
		})

		It("ignores that data", func() {
			Expect(currentUsage).To(BeEmpty())
		})
	})

	When("cache returns data for instances with unknown process instance id", func() {
		BeforeEach(func() {
			logCacheClient.PromQLReturns(queryResult(
				sample("0", "xyz",
					point("1", 0.5),
				),
			), nil)
		})

		It("ignores that data", func() {
			Expect(currentUsage).To(BeEmpty())
		})
	})

	When("fetching the current usage fails", func() {
		BeforeEach(func() {
			logCacheClient.PromQLReturns(nil, errors.New("fetch-failed"))
		})

		It("returns the error", func() {
			Expect(fetchErr).To(MatchError("fetch-failed"))
		})
	})

	When("fetched data has corrupt instance id", func() {
		BeforeEach(func() {
			logCacheClient.PromQLReturns(queryResult(
				sample("0", "abc",
					point("1", 0.2),
				),
				sample("1", "def",
					point("2", 0.3),
				),
				sample("dyado", "def",
					point("2", 0.4),
				),
				sample("2", "ghi",
					point("4", 0.5),
				),
			), nil)
		})

		It("ignores the corrupt data point", func() {
			Expect(fetchErr).NotTo(HaveOccurred())
			Expect(currentUsage).To(Equal(map[int][]fetchers.InstanceData{
				0: {
					{
						InstanceID: 0,
						Value:      0.2,
					},
				},
				1: {
					{
						InstanceID: 1,
						Value:      0.3,
					},
				},
				2: {
					{
						InstanceID: 2,
						Value:      0.5,
					},
				},
			}))
		})
	})
})
