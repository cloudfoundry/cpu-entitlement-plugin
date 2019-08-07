package fetchers_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"code.cloudfoundry.org/cpu-entitlement-plugin/fetchers"
	"code.cloudfoundry.org/cpu-entitlement-plugin/fetchers/fetchersfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CurrentUsage", func() {
	var (
		logCacheClient *fetchersfakes.FakeLogCacheClient
		fetcher        *fetchers.CurrentUsageFetcher
		appGuid        string
		currentUsage   map[int][]fetchers.InstanceData
		fetchErr       error
	)

	BeforeEach(func() {
		logCacheClient = new(fetchersfakes.FakeLogCacheClient)
		fetcher = fetchers.NewCurrentUsageFetcher(logCacheClient)

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
	})

	JustBeforeEach(func() {
		currentUsage, fetchErr = fetcher.FetchInstanceData(appGuid)
	})

	It("gets the current usage from the log-cache client", func() {
		Expect(logCacheClient.PromQLCallCount()).To(Equal(1))
		ctx, query, _ := logCacheClient.PromQLArgsForCall(0)
		Expect(ctx).To(Equal(context.Background()))
		Expect(query).To(Equal(fmt.Sprintf(`idelta(absolute_usage{source_id="%s"}[1m]) / idelta(absolute_entitlement{source_id="%s"}[1m])`, appGuid, appGuid)))
	})

	It("returns the correct current usage", func() {
		Expect(fetchErr).NotTo(HaveOccurred())
		Expect(currentUsage).To(Equal(map[int][]fetchers.InstanceData{
			0: {
				{
					Time:       time.Unix(1, 0),
					InstanceID: 0,
					Value:      0.2,
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

	When("fetching the current usage fails", func() {
		BeforeEach(func() {
			logCacheClient.PromQLReturns(nil, errors.New("fetch-failed"))
		})

		It("returns the error", func() {
			Expect(fetchErr).To(MatchError("fetch-failed"))
		})
	})

	When("fetched data has corrupt timestamp", func() {
		BeforeEach(func() {
			logCacheClient.PromQLReturns(queryResult(
				sample("0", "abc",
					point("1", 0.2),
				),
				sample("1", "def",
					point("baba", 0.4),
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
						Time:       time.Unix(1, 0),
						InstanceID: 0,
						Value:      0.2,
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

	When("fetched data has corrupt instance id", func() {
		BeforeEach(func() {
			logCacheClient.PromQLReturns(queryResult(
				sample("0", "abc",
					point("1", 0.2),
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
						Time:       time.Unix(1, 0),
						InstanceID: 0,
						Value:      0.2,
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
})
