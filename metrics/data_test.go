package metrics_test

import (
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cpu-entitlement-plugin/metrics"
)

var _ = Describe("Usagemetric", func() {
	Describe("FromGaugeMetric", func() {
		var (
			gaugeValues map[string]*loggregator_v2.GaugeValue
			ok          bool
			usageMetric metrics.InstanceData
		)

		BeforeEach(func() {
			gaugeValues = map[string]*loggregator_v2.GaugeValue{
				"absolute_usage":       &loggregator_v2.GaugeValue{Value: 1},
				"absolute_entitlement": &loggregator_v2.GaugeValue{Value: 2},
				"container_age":        &loggregator_v2.GaugeValue{Value: 3},
			}
		})

		JustBeforeEach(func() {
			usageMetric, ok = metrics.InstanceDataFromGauge("0", gaugeValues)
		})

		It("builds an InstanceData metric from a gauge metric message map", func() {
			Expect(ok).To(BeTrue())
			Expect(usageMetric).To(Equal(metrics.InstanceData{
				InstanceId:          0,
				AbsoluteUsage:       1,
				AbsoluteEntitlement: 2,
				ContainerAge:        3,
			}))
		})

		Context("when the gauce metric is missing the absolute_usage", func() {
			BeforeEach(func() {
				delete(gaugeValues, "absolute_usage")
			})

			It("returns !ok", func() {
				Expect(ok).To(BeFalse())
				Expect(usageMetric).To(Equal(metrics.InstanceData{}))
			})
		})

		Context("when the gauce metric is missing the absolute_entitlement", func() {
			BeforeEach(func() {
				delete(gaugeValues, "absolute_entitlement")
			})

			It("returns !ok", func() {
				Expect(ok).To(BeFalse())
				Expect(usageMetric).To(Equal(metrics.InstanceData{}))
			})
		})

		Context("when the gauce metric is missing the container_age", func() {
			BeforeEach(func() {
				delete(gaugeValues, "container_age")
			})

			It("returns !ok", func() {
				Expect(ok).To(BeFalse())
				Expect(usageMetric).To(Equal(metrics.InstanceData{}))
			})
		})
	})
})
