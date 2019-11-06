package integration_test

import (
	"code.cloudfoundry.org/cpu-entitlement-plugin/fetchers"
	"github.com/masters-of-cats/test-log-emitter/emitters"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ProcessInstanceIDFetcher", func() {

	var (
		appID   string
		fetcher fetchers.ProcessInstanceIDFetcher
	)

	BeforeEach(func() {
		uid := uuid.New().String()
		appID = "test-app-" + uid
		fetcher = fetchers.NewProcessInstanceIDFetcherWithLimit(logCacheClient, 2)

		emitUsage(appID, "3", "3a", 0, 0)
		emitUsage(appID, "1", "1a", 0, 0)
		emitUsage(appID, "1", "1b", 0, 0)
		emitUsage(appID, "2", "2a", 0, 0)
		emitUsage(appID, "3", "3b", 0, 0)
		emitCounter(emitters.CounterMetric{
			Name:       "absolute_entitlement",
			SourceId:   appID,
			InstanceId: "3",
			Tags: map[string]string{
				"process_instance_id": "3c",
			},
		})
		emitGauge(emitters.GaugeMetric{
			SourceId:   appID,
			InstanceId: "2",
			Tags:       map[string]string{"process_instance_id": "2b"},
			Values: []emitters.GaugeValue{
				{Name: "absolute_disaster", Value: 42, Unit: "nanoseconds"},
			},
		})
	})

	It("fetches the last process instance id for each instance", func() {
		Eventually(func() (map[int]string, error) { return fetcher.Fetch(appID) }, "15s").Should(Equal(map[int]string{
			1: "1b",
			2: "2a",
			3: "3b",
		}))
	})
})
