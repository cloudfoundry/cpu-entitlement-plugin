package fetchers_test

import (
	"errors"

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
		cumulativeUsage map[int]interface{}
		fetchErr        error
	)

	BeforeEach(func() {
		logCacheClient = new(fetchersfakes.FakeLogCacheClient)
		fetcher = fetchers.NewCumulativeUsageFetcher(logCacheClient)

		appGuid = "foo"

		appInstances = map[int]cf.Instance{
			0: cf.Instance{InstanceID: 0, ProcessInstanceID: "abc"},
		}

	})

	JustBeforeEach(func() {
		cumulativeUsage, fetchErr = fetcher.FetchInstanceData(appGuid, appInstances)
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
