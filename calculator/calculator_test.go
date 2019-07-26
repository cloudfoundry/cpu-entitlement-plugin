package calculator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cpu-entitlement-plugin/calculator"
	"code.cloudfoundry.org/cpu-entitlement-plugin/metrics"
)

var _ = Describe("Calculator", func() {
	var (
		usages []metrics.InstanceData
		calc   calculator.Calculator
		infos  []calculator.InstanceReport
	)

	BeforeEach(func() {
		usages = []metrics.InstanceData{
			{
				InstanceId:          0,
				AbsoluteUsage:       1.0,
				AbsoluteEntitlement: 2.0,
				ContainerAge:        3.0,
			},
			{
				InstanceId:          1,
				AbsoluteUsage:       4.0,
				AbsoluteEntitlement: 5.0,
				ContainerAge:        6.0,
			},
		}
		calc = calculator.New()
	})

	It("calculates entitlement ratio", func() {
		infos = calc.CalculateInstanceReports(usages)
		Expect(infos).To(Equal([]calculator.InstanceReport{
			{InstanceId: 0, EntitlementUsage: 0.5},
			{InstanceId: 1, EntitlementUsage: 0.8},
		}))
	})
})
