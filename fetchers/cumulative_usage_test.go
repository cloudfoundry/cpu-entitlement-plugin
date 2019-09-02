package fetchers_test

import (
	"context"
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cpu-entitlement-plugin/fetchers"
	"code.cloudfoundry.org/cpu-entitlement-plugin/fetchers/fetchersfakes"
)

var _ = Describe("Fetchers/CumulativeUsage", func() {
	var (
		logCacheClient  *fetchersfakes.FakeLogCacheClient
		fetcher         fetchers.CumulativeUsageFetcher
		appGuid         string
		cumulativeUsage []float64
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
	})

	JustBeforeEach(func() {
		cumulativeUsage, fetchErr = fetcher.FetchInstanceEntitlementUsages(appGuid)
	})

	It("gets the current usage from the log-cache client", func() {
		Expect(logCacheClient.PromQLCallCount()).To(Equal(1))
		ctx, query, _ := logCacheClient.PromQLArgsForCall(0)
		Expect(ctx).To(Equal(context.Background()))
		Expect(query).To(Equal(fmt.Sprintf(`absolute_usage{source_id="%s"} / absolute_entitlement{source_id="%s"}`, appGuid, appGuid)))
	})

	It("returns the correct accumulated usage", func() {
		Expect(fetchErr).NotTo(HaveOccurred())
		Expect(cumulativeUsage).To(ConsistOf(0.2, 0.4, 0.5))
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
