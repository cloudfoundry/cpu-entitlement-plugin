package metrics_test

import (
	"time"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cpu-entitlement-plugin/metrics"
)

var _ = Describe("InstanceData", func() {
	Describe("FromEnvelope", func() {
		var (
			envelope    loggregator_v2.Envelope
			gaugeValues map[string]*loggregator_v2.GaugeValue
			ok          bool
			data        metrics.InstanceData
		)

		BeforeEach(func() {
			gaugeValues = map[string]*loggregator_v2.GaugeValue{
				"absolute_usage":       &loggregator_v2.GaugeValue{Value: 1},
				"absolute_entitlement": &loggregator_v2.GaugeValue{Value: 2},
				"container_age":        &loggregator_v2.GaugeValue{Value: 3},
			}
		})

		JustBeforeEach(func() {
			envelope = loggregator_v2.Envelope{
				Timestamp:  123,
				InstanceId: "0",
				Message: &loggregator_v2.Envelope_Gauge{
					Gauge: &loggregator_v2.Gauge{
						Metrics: gaugeValues,
					},
				},
			}
			data, ok = metrics.InstanceDataFromEnvelope(envelope)
		})

		It("builds an InstanceData metric from a gauge metric message map", func() {
			Expect(ok).To(BeTrue())
			Expect(data).To(Equal(metrics.InstanceData{
				Time:                time.Unix(0, 123),
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
				Expect(data).To(Equal(metrics.InstanceData{}))
			})
		})

		Context("when the gauce metric is missing the absolute_entitlement", func() {
			BeforeEach(func() {
				delete(gaugeValues, "absolute_entitlement")
			})

			It("returns !ok", func() {
				Expect(ok).To(BeFalse())
				Expect(data).To(Equal(metrics.InstanceData{}))
			})
		})

		Context("when the gauce metric is missing the container_age", func() {
			BeforeEach(func() {
				delete(gaugeValues, "container_age")
			})

			It("returns !ok", func() {
				Expect(ok).To(BeFalse())
				Expect(data).To(Equal(metrics.InstanceData{}))
			})
		})
	})
})
