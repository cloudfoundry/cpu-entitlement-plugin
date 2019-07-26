package calculator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cpu-entitlement-plugin/calculator"
	"code.cloudfoundry.org/cpu-entitlement-plugin/metrics"
)

var _ = Describe("Calculator", func() {
	var (
		usages  []metrics.InstanceData
		calc    calculator.Calculator
		reports []calculator.InstanceReport
	)

	BeforeEach(func() {
		usages = []metrics.InstanceData{
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
		reports = calc.CalculateInstanceReports(usages)
	})

	It("calculates entitlement ratio", func() {
		Expect(reports).To(Equal([]calculator.InstanceReport{
			{InstanceId: 0, EntitlementUsage: 0.5},
			{InstanceId: 1, EntitlementUsage: 0.875},
		}))
	})

	When("an instance is missing from the data", func() {
		BeforeEach(func() {
			usages = []metrics.InstanceData{
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
})
