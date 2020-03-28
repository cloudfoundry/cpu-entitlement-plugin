package integration_test

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cpu-entitlement-plugin/cf"
	"code.cloudfoundry.org/cpu-entitlement-plugin/fetchers"
	"code.cloudfoundry.org/cpu-entitlement-plugin/httpclient"
	. "code.cloudfoundry.org/cpu-entitlement-plugin/test_utils"
	logcache "code.cloudfoundry.org/log-cache/pkg/client"
	"github.com/google/uuid"
	"github.com/masters-of-cats/test-log-emitter/emitters"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cumulative Usage Fetcher", func() {
	var (
		uid    string
		app1ID string
		app2ID string

		fetcher fetchers.CumulativeUsageFetcher
	)

	getUsages := func(appGUID string, appInstances map[int]cf.Instance) map[int]interface{} {
		usages, err := fetcher.FetchInstanceData(logger, appGUID, appInstances)
		Expect(err).NotTo(HaveOccurred())
		return usages
	}

	BeforeEach(func() {
		uid = uuid.New().String()
		app1ID = fmt.Sprintf("app-%s-1", uid)
		app2ID = fmt.Sprintf("app-%s-2", uid)

		fetcher = fetchers.NewCumulativeUsageFetcher(logCacheClient)
	})

	When("running multiple apps with various instance counts", func() {
		BeforeEach(func() {
			emitUsage(app1ID, "1", "1", 100, 150)
			emitUsage(app1ID, "1", "1", 134, 186)
			emitUsage(app1ID, "2", "2", 135, 137)
			emitUsage(app1ID, "3", "3", 136, 138)
			emitUsage(app2ID, "1", "10", 236, 238)
		})

		It("gets the usages of all instances for each app", func() {
			instances1 := map[int]cf.Instance{
				1: {InstanceID: 1, ProcessInstanceID: "1"},
				2: {InstanceID: 2, ProcessInstanceID: "2"},
				3: {InstanceID: 3, ProcessInstanceID: "3"},
			}
			instances2 := map[int]cf.Instance{
				1: {InstanceID: 1, ProcessInstanceID: "10"},
			}
			Eventually(func() map[int]interface{} { return getUsages(app1ID, instances1) }).Should(HaveLen(3))
			Eventually(func() map[int]interface{} { return getUsages(app2ID, instances2) }).Should(HaveLen(1))
		})

		It("gets the latest value when > 1 exists", func() {
			instances1 := map[int]cf.Instance{
				1: {InstanceID: 1, ProcessInstanceID: "1"},
			}
			var app1Usages map[int]interface{}
			Eventually(func() map[int]interface{} { app1Usages = getUsages(app1ID, instances1); return app1Usages }).Should(HaveLen(1))

			app1Inst1Usage, ok := app1Usages[1].(fetchers.CumulativeInstanceData)
			Expect(ok).To(BeTrue(), "couldn't cast fetcher result to CumulativeInstanceData")
			Expect(app1Inst1Usage.Usage).To(BeNumerically("~", 134.0/186.0, 0.00001))
		})
	})

	When("a metric exists for an old instance", func() {
		BeforeEach(func() {
			emitUsage(app1ID, "1", "1", 134, 186)
			emitUsage(app1ID, "1", "old", 50, 100)
		})

		It("is ignored, and an earlier correct instance value is returned", func() {
			instances1 := map[int]cf.Instance{
				1: {InstanceID: 1, ProcessInstanceID: "1"},
			}
			var app1Usages map[int]interface{}
			Eventually(func() map[int]interface{} { app1Usages = getUsages(app1ID, instances1); return app1Usages }).Should(HaveLen(1))

			app1Inst1Usage, ok := app1Usages[1].(fetchers.CumulativeInstanceData)
			Expect(ok).To(BeTrue(), "couldn't cast fetcher result to CumulativeInstanceData")
			Expect(app1Inst1Usage.Usage).To(BeNumerically("~", 134.0/186.0, 0.00001))
		})
	})

	When("the log-cache URL is not correct", func() {
		BeforeEach(func() {
			logCacheClient := logcache.NewClient(
				"http://1.2.3:123",
				logcache.WithHTTPClient(httpclient.NewAuthClient(getToken)),
			)

			fetcher = fetchers.NewCumulativeUsageFetcher(logCacheClient)
		})

		It("returns an error about the url", func() {
			_, err := fetcher.FetchInstanceData(logger, "anything", nil)
			Expect(err).To(MatchError(ContainSubstring("dial")))
		})
	})
})

func getCmdOutput(cmd string, args ...string) string {
	return strings.TrimSpace(string(Cmd(cmd, args...).Run().Out.Contents()))
}

func emitUsage(appID, instanceID, processInstanceID string, usage, entitlement float64) {
	emitGauge(emitters.GaugeMetric{
		SourceId:   appID,
		InstanceId: instanceID,
		Tags:       map[string]string{"process_instance_id": processInstanceID},
		Values: []emitters.GaugeValue{
			{Name: "absolute_usage", Value: usage, Unit: "nanoseconds"},
			{Name: "absolute_entitlement", Value: entitlement, Unit: "nanoseconds"},
		},
	})
}
