package calculator_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cpu-entitlement-plugin/calculator"
	"code.cloudfoundry.org/cpu-entitlement-plugin/metrics"
)

var _ = Describe("Calculator", func() {
	var (
		data    []metrics.InstanceData
		calc    calculator.Calculator
		reports []calculator.InstanceReport
	)

	BeforeEach(func() {
		data = []metrics.InstanceData{
			{
				InstanceId:          1,
				AbsoluteUsage:       7.0,
				AbsoluteEntitlement: 8.0,
				ContainerAge:        9.0,
			},
			{
				InstanceId:          1,
				AbsoluteUsage:       4.0,
				AbsoluteEntitlement: 5.0,
				ContainerAge:        6.0,
			},
			{
				InstanceId:          0,
				AbsoluteUsage:       1.0,
				AbsoluteEntitlement: 2.0,
				ContainerAge:        3.0,
			},
		}
		calc = calculator.New()
	})

	JustBeforeEach(func() {
		reports = calc.CalculateInstanceReports(data)
	})

	It("calculates entitlement ratio", func() {
		Expect(reports).To(Equal([]calculator.InstanceReport{
			{InstanceId: 0, EntitlementUsage: 0.5},
			{InstanceId: 1, EntitlementUsage: 0.875},
		}))
	})

	When("an instance is missing from the data", func() {
		BeforeEach(func() {
			data = []metrics.InstanceData{
				{
					InstanceId:          2,
					AbsoluteUsage:       4.0,
					AbsoluteEntitlement: 5.0,
					ContainerAge:        6.0,
				},
				{
					InstanceId:          0,
					AbsoluteUsage:       1.0,
					AbsoluteEntitlement: 2.0,
					ContainerAge:        3.0,
				},
			}
		})

		It("still returns an (incomplete) result", func() {
			Expect(reports).To(Equal([]calculator.InstanceReport{
				{InstanceId: 0, EntitlementUsage: 0.5},
				{InstanceId: 2, EntitlementUsage: 0.8},
			}))
		})
	})

	When("some instances have spiked", func() {
		BeforeEach(func() {
			data = []metrics.InstanceData{
				{InstanceId: 0, Time: time.Unix(6, 0), AbsoluteUsage: 95.0, AbsoluteEntitlement: 100.0},
				{InstanceId: 0, Time: time.Unix(5, 0), AbsoluteUsage: 90.0, AbsoluteEntitlement: 30.0},
				{InstanceId: 1, Time: time.Unix(4, 0), AbsoluteUsage: 20.0, AbsoluteEntitlement: 50.0},
				{InstanceId: 0, Time: time.Unix(3, 0), AbsoluteUsage: 50.0, AbsoluteEntitlement: 25.0},
				{InstanceId: 1, Time: time.Unix(2, 0), AbsoluteUsage: 10.0, AbsoluteEntitlement: 15.0},
				{InstanceId: 0, Time: time.Unix(1, 0), AbsoluteUsage: 10.0, AbsoluteEntitlement: 20.0},
			}
		})

		It("adds the spike starting and ending times to the report", func() {
			Expect(reports[0].LastSpikeFrom).To(Equal(time.Unix(3, 0)))
			Expect(reports[0].LastSpikeTo).To(Equal(time.Unix(5, 0)))
		})
	})

	When("latest spike starts at beginning of data and ends before end of data", func() {
		BeforeEach(func() {
			data = []metrics.InstanceData{
				{InstanceId: 0, Time: time.Unix(3, 0), AbsoluteUsage: 95.0, AbsoluteEntitlement: 100.0},
				{InstanceId: 0, Time: time.Unix(2, 0), AbsoluteUsage: 90.0, AbsoluteEntitlement: 40.0},
				{InstanceId: 0, Time: time.Unix(1, 0), AbsoluteUsage: 90.0, AbsoluteEntitlement: 30.0},
			}
		})

		It("reports spike from beginning of data to end of spike", func() {
			Expect(reports[0].LastSpikeFrom).To(Equal(time.Unix(1, 0)))
			Expect(reports[0].LastSpikeTo).To(Equal(time.Unix(2, 0)))
		})
	})

	When("latest spike starts at beginning of data and is always spiking in range", func() {
		BeforeEach(func() {
			data = []metrics.InstanceData{
				{InstanceId: 0, Time: time.Unix(2, 0), AbsoluteUsage: 90.0, AbsoluteEntitlement: 40.0},
				{InstanceId: 0, Time: time.Unix(1, 0), AbsoluteUsage: 90.0, AbsoluteEntitlement: 30.0},
			}
		})

		It("reports spike from beginning of data to end of data", func() {
			Expect(reports[0].LastSpikeFrom).To(Equal(time.Unix(1, 0)))
			Expect(reports[0].LastSpikeTo).To(Equal(time.Unix(2, 0)))
		})
	})

	When("latest spike is spiking at end of data", func() {
		BeforeEach(func() {
			data = []metrics.InstanceData{
				{InstanceId: 0, Time: time.Unix(3, 0), AbsoluteUsage: 90.0, AbsoluteEntitlement: 50.0},
				{InstanceId: 0, Time: time.Unix(2, 0), AbsoluteUsage: 50.0, AbsoluteEntitlement: 40.0},
				{InstanceId: 0, Time: time.Unix(1, 0), AbsoluteUsage: 20.0, AbsoluteEntitlement: 30.0},
			}
		})

		It("reports spike from beginning of spike to end of data", func() {
			Expect(reports[0].LastSpikeFrom).To(Equal(time.Unix(2, 0)))
			Expect(reports[0].LastSpikeTo).To(Equal(time.Unix(3, 0)))
		})
	})

	When("multiple spikes exist", func() {
		BeforeEach(func() {
			data = []metrics.InstanceData{
				{InstanceId: 0, Time: time.Unix(7, 0), AbsoluteUsage: 95.0, AbsoluteEntitlement: 90.0},
				{InstanceId: 0, Time: time.Unix(6, 0), AbsoluteUsage: 85.0, AbsoluteEntitlement: 80.0},
				{InstanceId: 0, Time: time.Unix(5, 0), AbsoluteUsage: 65.0, AbsoluteEntitlement: 70.0},
				{InstanceId: 0, Time: time.Unix(4, 0), AbsoluteUsage: 65.0, AbsoluteEntitlement: 60.0},
				{InstanceId: 0, Time: time.Unix(3, 0), AbsoluteUsage: 55.0, AbsoluteEntitlement: 50.0},
				{InstanceId: 0, Time: time.Unix(2, 0), AbsoluteUsage: 30.0, AbsoluteEntitlement: 40.0},
			}
		})

		It("reports only the latest spike", func() {
			Expect(reports[0].LastSpikeFrom).To(Equal(time.Unix(6, 0)))
			Expect(reports[0].LastSpikeTo).To(Equal(time.Unix(7, 0)))
		})
	})

	When("a spike consists of a single data point", func() {
		BeforeEach(func() {
			data = []metrics.InstanceData{
				{InstanceId: 0, Time: time.Unix(4, 0), AbsoluteUsage: 60.0, AbsoluteEntitlement: 70.0},
				{InstanceId: 0, Time: time.Unix(3, 0), AbsoluteUsage: 55.0, AbsoluteEntitlement: 50.0},
				{InstanceId: 0, Time: time.Unix(2, 0), AbsoluteUsage: 30.0, AbsoluteEntitlement: 40.0},
			}
		})

		It("reports an empty range", func() {
			Expect(reports[0].LastSpikeFrom).To(Equal(time.Unix(3, 0)))
			Expect(reports[0].LastSpikeTo).To(Equal(time.Unix(3, 0)))
		})
	})

	When("an instance reaches 100% entitlement usage but doesn't go above", func() {
		BeforeEach(func() {
			data = []metrics.InstanceData{
				{InstanceId: 0, Time: time.Unix(4, 0), AbsoluteUsage: 60.0, AbsoluteEntitlement: 70.0},
				{InstanceId: 0, Time: time.Unix(3, 0), AbsoluteUsage: 50.0, AbsoluteEntitlement: 50.0},
				{InstanceId: 0, Time: time.Unix(2, 0), AbsoluteUsage: 30.0, AbsoluteEntitlement: 40.0},
			}
		})

		It("does not report a spike", func() {
			Expect(reports[0].LastSpikeFrom.IsZero()).To(BeTrue())
			Expect(reports[0].LastSpikeTo.IsZero()).To(BeTrue())
		})
	})
})
