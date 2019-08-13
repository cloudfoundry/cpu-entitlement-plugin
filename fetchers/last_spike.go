package fetchers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"code.cloudfoundry.org/cpu-entitlement-plugin/metadata"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"code.cloudfoundry.org/log-cache/pkg/client"
	"code.cloudfoundry.org/log-cache/pkg/rpc/logcache_v1"
)

type LastSpikeFetcher struct {
	client LogCacheClient
	since  time.Time
}

func NewLastSpikeFetcher(client LogCacheClient, since time.Time) *LastSpikeFetcher {
	return &LastSpikeFetcher{client: client, since: since}
}

func (f LastSpikeFetcher) FetchInstanceData(appGUID string, appInstances map[int]metadata.CFAppInstance) (map[int][]InstanceData, error) {
	// ?envelope_type=GAUGE&name_filter=spike_change&limit=1&descending=true"
	res, err := f.client.Read(context.Background(), appGUID, f.since, client.WithEnvelopeTypes(logcache_v1.EnvelopeType_GAUGE), client.WithLimit(1), client.WithDescending(), client.WithNameFilter("spike"))
	if err != nil {
		return nil, err
	}

	return parseLastSpike(res, appInstances)
}

func parseLastSpike(res []*loggregator_v2.Envelope, appInstances map[int]metadata.CFAppInstance) (map[int][]InstanceData, error) {
	lastSpikePerInstance := map[int][]InstanceData{}
	for _, envelope := range res {
		instanceID, err := strconv.Atoi(envelope.InstanceId)
		if err != nil {
			return nil, err
		}

		envelopeGauge, ok := envelope.Message.(*loggregator_v2.Envelope_Gauge)
		if !ok {
			return nil, fmt.Errorf("envelope is not a gauge: %#v", envelope)
		}
		gaugeValues := envelopeGauge.Gauge.Metrics
		spikeStart := time.Unix(int64(gaugeValues["spike_start"].Value), 0)
		spikeEnd := time.Unix(int64(gaugeValues["spike_end"].Value), 0)

		dataPoint := InstanceData{
			InstanceID: instanceID,
			From:       spikeStart,
			To:         spikeEnd,
		}

		if isValidSpike(dataPoint, appInstances) {
			lastSpikePerInstance[instanceID] = []InstanceData{dataPoint}
		}
	}

	return lastSpikePerInstance, nil
}

func isValidSpike(dataPoint InstanceData, appInstances map[int]metadata.CFAppInstance) bool {
	instance, instanceExists := appInstances[dataPoint.InstanceID]
	return instanceExists && (dataPoint.From.After(instance.Since) || dataPoint.From.Equal(instance.Since)) && (dataPoint.To.After(instance.Since) || dataPoint.To.Equal(instance.Since))
}
