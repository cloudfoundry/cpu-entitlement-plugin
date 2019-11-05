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
		usages, err := fetcher.FetchInstanceData(appGUID, appInstances)
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
			app1inst1 := emitters.GaugeMetric{
				SourceId:   app1ID,
				InstanceId: "1",
				Tags: map[string]string{
					"process_instance_id": "1",
				},
				Values: []emitters.GaugeValue{
					{Name: "absolute_usage", Value: 134, Unit: "nanoseconds"},
					{Name: "absolute_entitlement", Value: 186, Unit: "nanoseconds"},
				},
			}

			app1inst2 := emitters.GaugeMetric{
				SourceId:   app1ID,
				InstanceId: "2",
				Tags: map[string]string{
					"process_instance_id": "2",
				},
				Values: []emitters.GaugeValue{
					{Name: "absolute_usage", Value: 135, Unit: "nanoseconds"},
					{Name: "absolute_entitlement", Value: 137, Unit: "nanoseconds"},
				},
			}

			app1inst3 := emitters.GaugeMetric{
				SourceId:   app1ID,
				InstanceId: "3",
				Tags: map[string]string{
					"process_instance_id": "3",
				},
				Values: []emitters.GaugeValue{
					{Name: "absolute_usage", Value: 136, Unit: "nanoseconds"},
					{Name: "absolute_entitlement", Value: 138, Unit: "nanoseconds"},
				},
			}

			app2inst1 := emitters.GaugeMetric{
				SourceId:   app2ID,
				InstanceId: "1",
				Tags: map[string]string{
					"process_instance_id": "10",
				},
				Values: []emitters.GaugeValue{
					{Name: "absolute_usage", Value: 236, Unit: "nanoseconds"},
					{Name: "absolute_entitlement", Value: 238, Unit: "nanoseconds"},
				},
			}

			emitGauge(app1inst1)
			emitGauge(app1inst2)
			emitGauge(app1inst3)
			emitGauge(app2inst1)
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
			app1Usages := getUsages(app1ID, instances1)
			app2Usages := getUsages(app2ID, instances2)
			Expect(app1Usages).To(HaveLen(3))
			Expect(app2Usages).To(HaveLen(1))

			app1Inst1Usage, ok := app1Usages[1].(fetchers.CumulativeInstanceData)
			Expect(ok).To(BeTrue(), "couldn't cast fetcher result to CumulativeInstanceData")
			Expect(app1Inst1Usage.Usage).To(BeNumerically("~", 0.72043, 0.00001))
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
			_, err := fetcher.FetchInstanceData("anything", nil)
			Expect(err).To(MatchError(ContainSubstring("dial")))
		})
	})
})

func getCmdOutput(cmd string, args ...string) string {
	return strings.TrimSpace(string(Cmd(cmd, args...).Run().Out.Contents()))
}
