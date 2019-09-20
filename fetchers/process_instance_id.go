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
