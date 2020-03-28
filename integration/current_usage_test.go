package integration_test

import (
	"time"

	"code.cloudfoundry.org/cpu-entitlement-plugin/cf"
	"code.cloudfoundry.org/cpu-entitlement-plugin/fetchers"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CurrentUsage", func() {

	var (
		appID   string
		fetcher fetchers.CurrentUsageFetcher
	)

	getCurrentUsage := func(appID string, instanceID int, processInstanceID string) func() float64 {
		return func() float64 {
			usages, err := fetcher.FetchInstanceData(logger, appID, map[int]cf.Instance{instanceID: {InstanceID: instanceID, ProcessInstanceID: processInstanceID}})
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			if len(usages) != 1 {
				return -1
			}
			usage, ok := usages[instanceID].(fetchers.CurrentInstanceData)
			ExpectWithOffset(1, ok).To(BeTrue(), "didn't get a currentInstanceData")
			return usage.Usage
		}
	}

	BeforeEach(func() {
		uid := uuid.New().String()
		appID = "test-app-" + uid
		fetcher = fetchers.NewCurrentUsageFetcher(logCacheClient)
	})

	When("there are two or more metrics per app and instance ID", func() {
		BeforeEach(func() {
			emitUsage(appID, "1", "1", 50, 100)
			// idelta requires some time separation
			time.Sleep(time.Second)
			emitUsage(appID, "1", "1", 100, 200)
			// idelta requires some time separation
			time.Sleep(time.Second)
			emitUsage(appID, "1", "1", 150, 250)
		})

		It("returns the ideltas ratio of the most recent two instance metrics", func() {
			Eventually(getCurrentUsage(appID, 1, "1")).Should(Equal(1.0))
		})

	})

	When("some instances have two or more metrics, others have only 1, and others have none", func() {
		BeforeEach(func() {
			emitUsage(appID, "1", "1", 100, 200)
			emitUsage(appID, "2", "2", 100, 1000)
			// idelta requires some time separation
			time.Sleep(time.Second)
			emitUsage(appID, "1", "1", 150, 250)
		})

		It("falls back to the cumulative fetcher for the single metric instances and omits instances with no metrics", func() {
			Eventually(getCurrentUsage(appID, 1, "1")).Should(Equal(1.0))
			Eventually(getCurrentUsage(appID, 2, "2")).Should(Equal(0.1))
			Consistently(getCurrentUsage(appID, 3, "3")).Should(Equal(float64(-1)))
		})
	})

	When("metrics with old process_instance_ids exist", func() {
		BeforeEach(func() {
			emitUsage(appID, "1", "1", 100, 200)
			emitUsage(appID, "1", "old", 100, 230)
			// idelta requires some time separation
			time.Sleep(time.Second)
			emitUsage(appID, "1", "1", 150, 250)
		})

		It("ignores them", func() {
			Eventually(getCurrentUsage(appID, 1, "1")).Should(Equal(1.0))
		})
	})
})
