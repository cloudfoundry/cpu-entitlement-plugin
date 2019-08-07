package fetchers_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cpu-entitlement-plugin/fetchers"
	"code.cloudfoundry.org/cpu-entitlement-plugin/fetchers/fetchersfakes"
)

var _ = Describe("HistoricalUsageFetcher", func() {
	var (
		logCacheClient  *fetchersfakes.FakeLogCacheClient
		fetcher         *fetchers.HistoricalUsageFetcher
		appGuid         string
		historicalUsage map[int][]fetchers.InstanceData
		fetchErr        error
		from, to        time.Time
	)

	BeforeEach(func() {
		appGuid = "foo"
		from = time.Now().Add(-time.Hour)
		to = time.Now()
		logCacheClient = new(fetchersfakes.FakeLogCacheClient)
		fetcher = fetchers.NewHistoricalUsageFetcher(logCacheClient, from, to)

		logCacheClient.PromQLReturns(instantQueryResult(
			sample("0", "abc", nil),
			sample("1", "def", nil),
			sample("2", "ghi", nil),
		), nil)
	})

	JustBeforeEach(func() {
		historicalUsage, fetchErr = fetcher.FetchInstanceData(appGuid)
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

		It("gets the historical usage from the log-cache client", func() {
			Expect(logCacheClient.PromQLCallCount()).To(Equal(1))
			ctx, query, _ := logCacheClient.PromQLArgsForCall(0)
			Expect(ctx).To(Equal(context.Background()))
			Expect(query).To(Equal(fmt.Sprintf(`absolute_usage{source_id="%s"}`, appGuid)))

			Expect(logCacheClient.PromQLRangeCallCount()).To(Equal(1))
			ctx, query, _ = logCacheClient.PromQLRangeArgsForCall(0)
			Expect(ctx).To(Equal(context.Background()))
			Expect(query).To(Equal(fmt.Sprintf(`absolute_usage{source_id="%s"} / absolute_entitlement{source_id="%s"}`, appGuid, appGuid)))
		})

		It("returns the correct historical usage", func() {
			Expect(fetchErr).NotTo(HaveOccurred())
			Expect(historicalUsage).To(Equal(map[int][]fetchers.InstanceData{
				0: {
					{
						Time:       time.Unix(1, 0),
						InstanceID: 0,
						Value:      0.2,
					},
					{
						Time:       time.Unix(3, 0),
						InstanceID: 0,
						Value:      0.5,
					},
				},
				1: {
					{
						Time:       time.Unix(2, 0),
						InstanceID: 1,
						Value:      0.4,
					},
				},
				2: {
					{
						Time:       time.Unix(4, 0),
						InstanceID: 2,
						Value:      0.5,
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
			Expect(fetchErr).To(MatchError("boo"))
			Expect(historicalUsage).To(BeNil())
		})
	})

	When("fetching the list of data points from log-cache fails", func() {
		BeforeEach(func() {
			logCacheClient.PromQLRangeReturns(nil, errors.New("boo"))
		})

		It("returns an error", func() {
			Expect(fetchErr).To(MatchError("boo"))
			Expect(historicalUsage).To(BeNil())
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
			Expect(historicalUsage).To(Equal(map[int][]fetchers.InstanceData{
				0: {
					{
						Time:       time.Unix(1, 0),
						InstanceID: 0,
						Value:      0.5,
					},
				},
				1: {
					{
						Time:       time.Unix(2, 0),
						InstanceID: 1,
						Value:      0.5,
					},
				},
				2: {
					{
						Time:       time.Unix(3, 0),
						InstanceID: 2,
						Value:      0.5,
					},
				},
			}))
		})
	})
})
