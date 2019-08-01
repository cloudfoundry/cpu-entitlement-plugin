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
		data    map[int][]metrics.InstanceData
		calc    calculator.Calculator
		reports []calculator.InstanceReport
	)

	BeforeEach(func() {
		data = map[int][]metrics.InstanceData{
			0: {
				{
					InstanceID:       0,
					EntitlementUsage: 0.5,
				},
			},
			1: {
				{
					InstanceID:       1,
					EntitlementUsage: 0.6,
				},
				{
					InstanceID:       1,
					EntitlementUsage: 0.7,
				},
			},
		}
		calc = calculator.New()
	})

	JustBeforeEach(func() {
		reports = calc.CalculateInstanceReports(data)
	})

	It("calculates entitlement ratio", func() {
		Expect(reports).To(Equal([]calculator.InstanceReport{
			{InstanceID: 0, EntitlementUsage: 0.5},
			{InstanceID: 1, EntitlementUsage: 0.7},
		}))
	})

	When("an instance is missing from the data", func() {
		BeforeEach(func() {
			data = map[int][]metrics.InstanceData{
				2: {
					{
						InstanceID:       2,
						EntitlementUsage: 0.5,
					},
				},
				0: {
					{
						InstanceID:       0,
						EntitlementUsage: 0.6,
					},
				},
			}
		})

		It("still returns an (incomplete) result", func() {
			Expect(reports).To(Equal([]calculator.InstanceReport{
				{InstanceID: 0, EntitlementUsage: 0.6},
				{InstanceID: 2, EntitlementUsage: 0.5},
			}))
		})
	})

	When("some instances have spiked", func() {
		BeforeEach(func() {
			data = map[int][]metrics.InstanceData{
				0: {
					{InstanceID: 0, Time: time.Unix(1, 0), EntitlementUsage: 0.5},
					{InstanceID: 0, Time: time.Unix(3, 0), EntitlementUsage: 1.5},
					{InstanceID: 0, Time: time.Unix(5, 0), EntitlementUsage: 2.0},
					{InstanceID: 0, Time: time.Unix(6, 0), EntitlementUsage: 0.9},
				},
				1: {
					{InstanceID: 1, Time: time.Unix(2, 0), EntitlementUsage: 0.6},
					{InstanceID: 1, Time: time.Unix(4, 0), EntitlementUsage: 0.4},
				},
			}
		})

		It("adds the spike starting and ending times to the report", func() {
			Expect(reports[0].LastSpikeFrom).To(Equal(time.Unix(3, 0)))
			Expect(reports[0].LastSpikeTo).To(Equal(time.Unix(5, 0)))
		})
	})

	When("latest spike starts at beginning of data and ends before end of data", func() {
		BeforeEach(func() {
			data = map[int][]metrics.InstanceData{
				0: {
					{InstanceID: 0, Time: time.Unix(1, 0), EntitlementUsage: 2.5},
					{InstanceID: 0, Time: time.Unix(2, 0), EntitlementUsage: 1.5},
					{InstanceID: 0, Time: time.Unix(3, 0), EntitlementUsage: 0.9},
				},
			}
		})

		It("reports spike from beginning of data to end of spike", func() {
			Expect(reports[0].LastSpikeFrom).To(Equal(time.Unix(1, 0)))
			Expect(reports[0].LastSpikeTo).To(Equal(time.Unix(2, 0)))
		})
	})

	When("latest spike starts at beginning of data and is always spiking in range", func() {
		BeforeEach(func() {
			data = map[int][]metrics.InstanceData{
				0: {
					{InstanceID: 0, Time: time.Unix(1, 0), EntitlementUsage: 1.5},
					{InstanceID: 0, Time: time.Unix(2, 0), EntitlementUsage: 2.5},
				},
			}
		})

		It("reports spike from beginning of data to end of data", func() {
			Expect(reports[0].LastSpikeFrom).To(Equal(time.Unix(1, 0)))
			Expect(reports[0].LastSpikeTo).To(Equal(time.Unix(2, 0)))
		})
	})

	When("latest spike is spiking at end of data", func() {
		BeforeEach(func() {
			data = map[int][]metrics.InstanceData{
				0: {
					{InstanceID: 0, Time: time.Unix(1, 0), EntitlementUsage: 0.5},
					{InstanceID: 0, Time: time.Unix(2, 0), EntitlementUsage: 1.5},
					{InstanceID: 0, Time: time.Unix(3, 0), EntitlementUsage: 2.5},
				},
			}
		})

		It("reports spike from beginning of spike to end of data", func() {
			Expect(reports[0].LastSpikeFrom).To(Equal(time.Unix(2, 0)))
			Expect(reports[0].LastSpikeTo).To(Equal(time.Unix(3, 0)))
		})
	})

	When("multiple spikes exist", func() {
		BeforeEach(func() {
			data = map[int][]metrics.InstanceData{
				0: {
					{InstanceID: 0, Time: time.Unix(2, 0), EntitlementUsage: 0.5},
					{InstanceID: 0, Time: time.Unix(3, 0), EntitlementUsage: 0.7},
					{InstanceID: 0, Time: time.Unix(4, 0), EntitlementUsage: 0.9},
					{InstanceID: 0, Time: time.Unix(5, 0), EntitlementUsage: 0.8},
					{InstanceID: 0, Time: time.Unix(6, 0), EntitlementUsage: 1.2},
					{InstanceID: 0, Time: time.Unix(7, 0), EntitlementUsage: 1.5},
				},
			}
		})

		It("reports only the latest spike", func() {
			Expect(reports[0].LastSpikeFrom).To(Equal(time.Unix(6, 0)))
			Expect(reports[0].LastSpikeTo).To(Equal(time.Unix(7, 0)))
		})
	})

	When("a spike consists of a single data point", func() {
		BeforeEach(func() {
			data = map[int][]metrics.InstanceData{
				0: {
					{InstanceID: 0, Time: time.Unix(2, 0), EntitlementUsage: 0.8},
					{InstanceID: 0, Time: time.Unix(3, 0), EntitlementUsage: 1.5},
					{InstanceID: 0, Time: time.Unix(4, 0), EntitlementUsage: 0.5},
				},
			}
		})

		It("reports an empty range", func() {
			Expect(reports[0].LastSpikeFrom).To(Equal(time.Unix(3, 0)))
			Expect(reports[0].LastSpikeTo).To(Equal(time.Unix(3, 0)))
		})
	})

	When("an instance reaches 100% entitlement usage but doesn't go above", func() {
		BeforeEach(func() {
			data = map[int][]metrics.InstanceData{
				0: {
					{InstanceID: 0, Time: time.Unix(2, 0), EntitlementUsage: 0.5},
					{InstanceID: 0, Time: time.Unix(3, 0), EntitlementUsage: 1.0},
					{InstanceID: 0, Time: time.Unix(4, 0), EntitlementUsage: 0.8},
				},
			}
		})

		It("does not report a spike", func() {
			Expect(reports[0].LastSpikeFrom.IsZero()).To(BeTrue())
			Expect(reports[0].LastSpikeTo.IsZero()).To(BeTrue())
		})
	})
})
