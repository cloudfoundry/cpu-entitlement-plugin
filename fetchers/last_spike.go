package fetchers

import (
	"context"
	"strconv"
	"time"

	"code.cloudfoundry.org/cpu-entitlement-plugin/cf"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	logcache "code.cloudfoundry.org/log-cache/pkg/client"
	"code.cloudfoundry.org/log-cache/pkg/rpc/logcache_v1"
)

type LastSpikeInstanceData struct {
	InstanceID int
	From       time.Time
	To         time.Time
}

type LastSpikeFetcher struct {
	client LogCacheClient
	since  time.Time
}

func NewLastSpikeFetcher(client LogCacheClient, since time.Time) *LastSpikeFetcher {
	return &LastSpikeFetcher{client: client, since: since}
}

func (f LastSpikeFetcher) FetchInstanceData(appGUID string, appInstances map[int]cf.Instance) (map[int]interface{}, error) {
	res, err := f.client.Read(context.Background(), appGUID, f.since,
		logcache.WithEnvelopeTypes(logcache_v1.EnvelopeType_GAUGE),
		logcache.WithDescending(),
		logcache.WithNameFilter("spike"),
	)
	if err != nil {
		return nil, err
	}

	return parseLastSpike(res, appInstances)
}

func parseLastSpike(res []*loggregator_v2.Envelope, appInstances map[int]cf.Instance) (map[int]interface{}, error) {
	lastSpikePerInstance := make(map[int]interface{})
	for _, envelope := range res {
		instanceID, err := strconv.Atoi(envelope.InstanceId)
		if err != nil {
			continue
		}

		if _, alreadySet := lastSpikePerInstance[instanceID]; alreadySet {
			continue
		}

		envelopeGauge, ok := envelope.Message.(*loggregator_v2.Envelope_Gauge)
		if !ok {
			continue
		}

		processInstanceID := envelope.Tags["process_instance_id"]
		if appInstances[instanceID].ProcessInstanceID != processInstanceID {
			continue
		}

		gaugeValues := envelopeGauge.Gauge.Metrics
		spikeStart := time.Unix(int64(gaugeValues["spike_start"].Value), 0)
		spikeEnd := time.Unix(int64(gaugeValues["spike_end"].Value), 0)

		lastSpikePerInstance[instanceID] = LastSpikeInstanceData{
			InstanceID: instanceID,
			From:       spikeStart,
			To:         spikeEnd,
		}
	}

	return lastSpikePerInstance, nil
}
