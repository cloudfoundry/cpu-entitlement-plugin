package calculator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cpu-entitlement-plugin/calculator"
	"code.cloudfoundry.org/cpu-entitlement-plugin/metrics"
)

var _ = Describe("Calculator", func() {
	var (
		usages []metrics.Usage
		calc   calculator.Calculator
		infos  []calculator.InstanceInfo
	)

	BeforeEach(func() {
		usages = []metrics.Usage{
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
		infos = calc.CalculateInstanceInfos(usages)
		Expect(infos).To(Equal([]calculator.InstanceInfo{
			{InstanceId: 0, EntitlementUsage: 0.5},
			{InstanceId: 1, EntitlementUsage: 0.8},
		}))
	})
})
