package metrics // import "code.cloudfoundry.org/cpu-entitlement-plugin/metrics"

import (
	"strconv"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
)

type gaugeMetric map[string]*loggregator_v2.GaugeValue

type InstanceData struct {
	InstanceId          int
	AbsoluteUsage       float64
	AbsoluteEntitlement float64
	ContainerAge        float64
}

func InstanceDataFromGauge(instanceId string, metric gaugeMetric) (InstanceData, bool) {
	absoluteUsage := metric["absolute_usage"]
	absoluteEntitlement := metric["absolute_entitlement"]
	containerAge := metric["container_age"]

	if absoluteUsage == nil {
		return InstanceData{}, false
	}

	if absoluteEntitlement == nil {
		return InstanceData{}, false
	}

	if containerAge == nil {
		return InstanceData{}, false
	}

	instanceIndex, err := strconv.Atoi(instanceId)
	if err != nil {
		return InstanceData{}, false
	}

	return InstanceData{
		InstanceId:          instanceIndex,
		AbsoluteUsage:       absoluteUsage.Value,
		AbsoluteEntitlement: absoluteEntitlement.Value,
		ContainerAge:        containerAge.Value,
	}, true
}
