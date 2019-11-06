package fetchers

import (
	"context"
	"strconv"
	"time"

	logcache "code.cloudfoundry.org/log-cache/pkg/client"
	"code.cloudfoundry.org/log-cache/pkg/rpc/logcache_v1"
)

const maxReadTries = 10

type ProcessInstanceIDFetcher struct {
	client LogCacheClient
	limit  int
}

func NewProcessInstanceIDFetcherWithLimit(client LogCacheClient, limit int) ProcessInstanceIDFetcher {
	return ProcessInstanceIDFetcher{
		client: client,
		limit:  limit,
	}
}

func NewProcessInstanceIDFetcher(client LogCacheClient) ProcessInstanceIDFetcher {
	return NewProcessInstanceIDFetcherWithLimit(client, 1000)
}

// Fetch searches in a 30s interval, in which each app instance will have
// emitted at least one metric. As log-cache read is limited to 1000 results,
// we have implemented some pagination here. We start with the topmost 1000
// results, and if we do receive 1000 back, there may be more, so we reduce
// `end` back to the timestamp of the earliest metric from the results and
// re-read. As soon as fewer than 1000 results are returned we stop, as we have
// exhausted the range. We also apply a 10 iteration sanity check to avoid
// looping forever.
func (f ProcessInstanceIDFetcher) Fetch(appGUID string) (map[int]string, error) {
	start := time.Now().Add(-30 * time.Second)
	end := time.Now()

	processInstanceIDs := map[int]string{}

	for i := 0; i < maxReadTries; i++ {
		envelopes, err := f.client.Read(context.Background(), appGUID, start,
			logcache.WithDescending(),
			logcache.WithEnvelopeTypes(logcache_v1.EnvelopeType_GAUGE),
			logcache.WithNameFilter("absolute_entitlement"),
			logcache.WithEndTime(end),
			logcache.WithLimit(f.limit),
		)

		if err != nil {
			return nil, err
		}

		for _, envelope := range envelopes {
			instanceID, err := strconv.Atoi(envelope.InstanceId)
			if err != nil {
				continue
			}

			processInstanceID := envelope.Tags["process_instance_id"]
			if len(processInstanceID) == 0 {
				continue
			}

			if _, exists := processInstanceIDs[instanceID]; !exists {
				processInstanceIDs[instanceID] = processInstanceID
			}
		}

		if len(envelopes) < f.limit {
			break
		}

		end = time.Unix(0, envelopes[len(envelopes)-1].Timestamp)
	}

	return processInstanceIDs, nil
}
