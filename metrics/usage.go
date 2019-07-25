package metrics // import "code.cloudfoundry.org/cpu-entitlement-plugin/metrics"

import (
	"strconv"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
)

type gaugeMetric map[string]*loggregator_v2.GaugeValue

type Usage struct {
	InstanceId          int
	AbsoluteUsage       float64
	AbsoluteEntitlement float64
	ContainerAge        float64
}

func UsageFromGauge(instanceId string, metric gaugeMetric) (Usage, bool) {
	absoluteUsage := metric["absolute_usage"]
	absoluteEntitlement := metric["absolute_entitlement"]
	containerAge := metric["container_age"]

	if absoluteUsage == nil {
		return Usage{}, false
	}

	if absoluteEntitlement == nil {
		return Usage{}, false
	}

	if containerAge == nil {
		return Usage{}, false
	}

	instanceIndex, err := strconv.Atoi(instanceId)
	if err != nil {
		return Usage{}, false
	}

	return Usage{
		InstanceId:          instanceIndex,
		AbsoluteUsage:       absoluteUsage.Value,
		AbsoluteEntitlement: absoluteEntitlement.Value,
		ContainerAge:        containerAge.Value,
	}, true
}

func (m Usage) EntitlementRatio() float64 {
	return m.AbsoluteUsage / m.AbsoluteEntitlement
}
