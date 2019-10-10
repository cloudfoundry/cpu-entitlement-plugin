package fetchers_test

import (
	"context"
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cpu-entitlement-plugin/cf"
	"code.cloudfoundry.org/cpu-entitlement-plugin/fetchers"
	"code.cloudfoundry.org/cpu-entitlement-plugin/fetchers/fetchersfakes"
)

var _ = Describe("Fetchers/CumulativeUsage", func() {
	var (
		logCacheClient  *fetchersfakes.FakeLogCacheClient
		fetcher         fetchers.CumulativeUsageFetcher
		appGuid         string
		appInstances    map[int]cf.Instance
		cumulativeUsage map[int][]fetchers.InstanceData
		fetchErr        error
	)

	BeforeEach(func() {
		logCacheClient = new(fetchersfakes.FakeLogCacheClient)
		fetcher = fetchers.NewCumulativeUsageFetcher(logCacheClient)

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
		cumulativeUsage, fetchErr = fetcher.FetchInstanceData(appGuid, appInstances)
	})

	It("gets the current usage from the log-cache client", func() {
		Expect(logCacheClient.PromQLCallCount()).To(Equal(1))
		ctx, query, _ := logCacheClient.PromQLArgsForCall(0)
		Expect(ctx).To(Equal(context.Background()))
		Expect(query).To(Equal(fmt.Sprintf(`absolute_usage{source_id="%s"} / absolute_entitlement{source_id="%s"}`, appGuid, appGuid)))
	})

	It("returns the correct accumulated usage", func() {
		Expect(fetchErr).NotTo(HaveOccurred())
		Expect(cumulativeUsage).To(SatisfyAll(HaveLen(3), HaveKey(0), HaveKey(1), HaveKey(2)))
		Expect(cumulativeUsage[0][0].Value).To(Equal(0.2))
		Expect(cumulativeUsage[1][0].Value).To(Equal(0.4))
		Expect(cumulativeUsage[2][0].Value).To(Equal(0.5))
	})

	When("cache returns data for instances that are no longer running (because the app has been scaled down", func() {
		BeforeEach(func() {
			appInstances = map[int]cf.Instance{
				0: cf.Instance{InstanceID: 0, ProcessInstanceID: "abc"},
			}
		})

		It("returns current usage for running instances only", func() {
			Expect(fetchErr).NotTo(HaveOccurred())
			Expect(cumulativeUsage).To(SatisfyAll(HaveLen(1), HaveKey(0)))
			Expect(cumulativeUsage[0][0].Value).To(Equal(0.2))
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
			Expect(cumulativeUsage).To(BeEmpty())
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
			Expect(cumulativeUsage).To(BeEmpty())
		})
	})

	When("fetched data has corrupt instance id", func() {
		BeforeEach(func() {
			logCacheClient.PromQLReturns(queryResult(
				sample("dyado", "def",
					point("2", 0.4),
				),
			), nil)
		})

		It("ignores the corrupt data point", func() {
			Expect(fetchErr).NotTo(HaveOccurred())
			Expect(cumulativeUsage).To(BeEmpty())
		})
	})

	When("fetching the cumulative usage fails", func() {
		BeforeEach(func() {
			logCacheClient.PromQLReturns(nil, errors.New("fetch-failed"))
		})

		It("returns the error", func() {
			Expect(fetchErr).To(MatchError("fetch-failed"))
		})
	})
})
