package calculator

import (
	"code.cloudfoundry.org/cpu-entitlement-plugin/metrics"
)

type Calculator struct{}

type InstanceReport struct {
	InstanceId       int
	EntitlementUsage float64
}

func New() Calculator {
	return Calculator{}
}

func (c Calculator) CalculateInstanceReports(usages []metrics.InstanceData) []InstanceReport {
	var infos []InstanceReport

	for _, usage := range usages {
		infos = append(infos, c.calculateInstanceReport(usage))
	}

	return infos
}

func (c Calculator) calculateInstanceReport(usage metrics.InstanceData) InstanceReport {
	return InstanceReport{
		InstanceId:       usage.InstanceId,
		EntitlementUsage: usage.AbsoluteUsage / usage.AbsoluteEntitlement,
	}
}
