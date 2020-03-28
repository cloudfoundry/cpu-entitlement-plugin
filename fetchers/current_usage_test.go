package fetchers_test

import (
	"errors"

	"code.cloudfoundry.org/cpu-entitlement-plugin/cf"
	"code.cloudfoundry.org/cpu-entitlement-plugin/fetchers"
	"code.cloudfoundry.org/cpu-entitlement-plugin/fetchers/fetchersfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("CurrentUsage", func() {
	var (
		logCacheClient      *fetchersfakes.FakeLogCacheClient
		fakeFallbackFetcher *fetchersfakes.FakeFetcher
		fetcher             fetchers.CurrentUsageFetcher
		appGuid             string
		appInstances        map[int]cf.Instance
		currentUsage        map[int]interface{}
		fetchErr            error
	)

	BeforeEach(func() {
		logCacheClient = new(fetchersfakes.FakeLogCacheClient)
		fakeFallbackFetcher = new(fetchersfakes.FakeFetcher)
		fetcher = fetchers.NewCurrentUsageFetcherWithFallbackFetcher(logCacheClient, fakeFallbackFetcher)

		appGuid = "foo"

		appInstances = map[int]cf.Instance{
			0: cf.Instance{InstanceID: 0, ProcessInstanceID: "abc"},
			1: cf.Instance{InstanceID: 1, ProcessInstanceID: "def"},
			2: cf.Instance{InstanceID: 2, ProcessInstanceID: "ghi"},
		}
	})

	JustBeforeEach(func() {
		currentUsage, fetchErr = fetcher.FetchInstanceData(logger, appGuid, appInstances)
	})

	It("logs start and end", func() {
		Expect(logger).To(gbytes.Say("current-usage-fetcher.start"))
		Expect(logger).To(gbytes.Say("current-usage-fetcher.end"))
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
			Expect(currentUsage).To(Equal(map[int]interface{}{
				0: fetchers.CurrentInstanceData{
					InstanceID: 0,
					Usage:      0.2,
				},
				1: fetchers.CurrentInstanceData{
					InstanceID: 1,
					Usage:      0.3,
				},
				2: fetchers.CurrentInstanceData{
					InstanceID: 2,
					Usage:      0.5,
				},
			}))
		})
	})
})
