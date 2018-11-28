package usagemetric // import "code.cloudfoundry.org/cpu-entitlement-plugin/usagemetric"

import "code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"

type gaugeMetric map[string]*loggregator_v2.GaugeValue

type UsageMetric struct {
	InstanceId          string
	AbsoluteUsage       float64
	AbsoluteEntitlement float64
	ContainerAge        float64
}

func FromGaugeMetric(instanceId string, metric gaugeMetric) (UsageMetric, bool) {
	absoluteUsage := metric["absolute_usage"]
	absoluteEntitlement := metric["absolute_entitlement"]
	containerAge := metric["container_age"]

	if absoluteUsage == nil {
		return UsageMetric{}, false
	}

	if absoluteEntitlement == nil {
		return UsageMetric{}, false
	}

	if containerAge == nil {
		return UsageMetric{}, false
	}

	return UsageMetric{
		InstanceId:          instanceId,
		AbsoluteUsage:       absoluteUsage.Value,
		AbsoluteEntitlement: absoluteEntitlement.Value,
		ContainerAge:        containerAge.Value,
	}, true
}

func (m UsageMetric) CPUUsage() float64 {
	return m.AbsoluteUsage / m.AbsoluteEntitlement
}
