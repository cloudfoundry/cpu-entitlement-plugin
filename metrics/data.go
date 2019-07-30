package metrics // import "code.cloudfoundry.org/cpu-entitlement-plugin/metrics"

import (
	"strconv"
	"time"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
)

type gaugeMetric map[string]*loggregator_v2.GaugeValue

type InstanceData struct {
	Time                time.Time
	InstanceId          int
	AbsoluteUsage       float64
	AbsoluteEntitlement float64
	ContainerAge        float64
}

func InstanceDataFromEnvelope(envelope loggregator_v2.Envelope) (InstanceData, bool) {
	metric := envelope.GetGauge().GetMetrics()
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

	instanceIndex, err := strconv.Atoi(envelope.GetInstanceId())
	if err != nil {
		return InstanceData{}, false
	}

	return InstanceData{
		Time:                time.Unix(0, envelope.Timestamp),
		InstanceId:          instanceIndex,
		AbsoluteUsage:       absoluteUsage.Value,
		AbsoluteEntitlement: absoluteEntitlement.Value,
		ContainerAge:        containerAge.Value,
	}, true
}
