package fetchers_test

import (
	"errors"

	"code.cloudfoundry.org/cpu-entitlement-plugin/fetchers"
	"code.cloudfoundry.org/cpu-entitlement-plugin/fetchers/fetchersfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ProcessInstanceId Fetcher", func() {

	var (
		appGUID            string
		logCacheClient     *fetchersfakes.FakeLogCacheClient
		fetcher            fetchers.ProcessInstanceIDFetcher
		processInstanceIDs map[int]string
		err                error
	)

	BeforeEach(func() {
		appGUID = "my-app-guid"
		logCacheClient = new(fetchersfakes.FakeLogCacheClient)
		logCacheClient.PromQLRangeReturns(rangeQueryResult(
			series("0", "abc",
				point("1.1", 0.2),
				point("2.2", 0.5),
			),
			series("1", "bcd",
				point("1.3", 0.2),
				point("2", 0.5),
			),
		), nil)
		fetcher = fetchers.NewProcessInstanceIDFetcher(logCacheClient)
	})

	JustBeforeEach(func() {
		processInstanceIDs, err = fetcher.Fetch(appGUID)
	})

	It("returns the process instance id map", func() {
		Expect(err).NotTo(HaveOccurred())
		Expect(processInstanceIDs).To(HaveKeyWithValue(0, "abc"))
		Expect(processInstanceIDs).To(HaveKeyWithValue(1, "bcd"))
		Expect(processInstanceIDs).To(HaveLen(2))

		Expect(logCacheClient.PromQLRangeCallCount()).To(Equal(1))
		_, actualQuery, _ := logCacheClient.PromQLRangeArgsForCall(0)
		Expect(actualQuery).To(ContainSubstring("source_id=\"my-app-guid\""))
	})

	When("there are duplicate series for an instance id", func() {
		BeforeEach(func() {
			logCacheClient.PromQLRangeReturns(rangeQueryResult(
				series("0", "abc",
					point("2", 0.2),
					point("3", 0.5),
				),
				series("0", "def",
					point("3", 0.4),
				),
				series("0", "ghi",
					point("1", 0.2),
					point("3", 0.5),
				),
			), nil)
		})

		It("returns the series with the most recent first data point", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(processInstanceIDs).To(HaveKeyWithValue(0, "def"))
			Expect(processInstanceIDs).To(HaveLen(1))
		})
	})

	When("a series has no points", func() {
		BeforeEach(func() {
			logCacheClient.PromQLRangeReturns(rangeQueryResult(series("0", "abc")), nil)
		})

		It("is ignored", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(processInstanceIDs).To(BeEmpty())
		})
	})

	When("a series has a nonsense instance_id", func() {
		BeforeEach(func() {
			logCacheClient.PromQLRangeReturns(rangeQueryResult(series("blah", "abc")), nil)
		})

		It("is ignored", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(processInstanceIDs).To(BeEmpty())
		})
	})

	When("a series has first point with a nonsense timestamp", func() {
		BeforeEach(func() {
			logCacheClient.PromQLRangeReturns(rangeQueryResult(series("0", "abc", point("bcd", 0.1))), nil)
		})

		It("is ignored", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(processInstanceIDs).To(BeEmpty())
		})
	})

	When("promql call returns an error", func() {
		BeforeEach(func() {
			logCacheClient.PromQLRangeReturns(nil, errors.New("log-cache-error"))
		})

		It("returns the error", func() {
			Expect(err).To(MatchError("log-cache-error"))
		})

	})
})
