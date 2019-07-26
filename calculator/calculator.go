package calculator

import (
	"code.cloudfoundry.org/cpu-entitlement-plugin/metrics"
)

type Calculator struct{}

type InstanceInfo struct {
	InstanceId       int
	EntitlementUsage float64
}

func New() Calculator {
	return Calculator{}
}

func (c Calculator) CalculateInstanceInfos(usages []metrics.Usage) []InstanceInfo {
	var infos []InstanceInfo

	for _, usage := range usages {
		infos = append(infos, c.calculateInstanceInfo(usage))
	}

	return infos
}

func (c Calculator) calculateInstanceInfo(usage metrics.Usage) InstanceInfo {
	return InstanceInfo{
		InstanceId:       usage.InstanceId,
		EntitlementUsage: usage.AbsoluteUsage / usage.AbsoluteEntitlement,
	}
}
