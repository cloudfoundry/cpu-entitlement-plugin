package fetchers

import (
	"context"
	"fmt"
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
	fmt.Printf("Response size: %d\n", len(res))
	for _, envelope := range res {
		instanceID, err := strconv.Atoi(envelope.InstanceId)
		if err != nil {
			fmt.Println("failed to parse instance id")
			continue
		}
		if _, set := lastSpikePerInstance[instanceID]; set {
			continue
		}

		envelopeGauge, ok := envelope.Message.(*loggregator_v2.Envelope_Gauge)
		if !ok {
			fmt.Println("this is not an envelope gauge")
			continue
		}
		gaugeValues := envelopeGauge.Gauge.Metrics
		spikeStart := time.Unix(int64(gaugeValues["spike_start"].Value), 0)
		spikeEnd := time.Unix(int64(gaugeValues["spike_end"].Value), 0)

		processInstanceID := envelope.Tags["process_instance_id"]
		if appInstances[instanceID].ProcessInstanceID != processInstanceID {
			fmt.Printf("process_instance_id: %q, expected_process_instance_id: %q\n", appInstances[instanceID].ProcessInstanceID, processInstanceID)
			continue
		}

		lastSpikePerInstance[instanceID] = LastSpikeInstanceData{
			InstanceID: instanceID,
			From:       spikeStart,
			To:         spikeEnd,
		}
	}

	return lastSpikePerInstance, nil
}
