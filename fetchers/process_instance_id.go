package fetchers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	logcache "code.cloudfoundry.org/log-cache/pkg/client"
)

type ProcessInstanceIDFetcher struct {
	client LogCacheClient
}

func NewProcessInstanceIDFetcher(logCacheClient LogCacheClient) ProcessInstanceIDFetcher {
	return ProcessInstanceIDFetcher{
		client: logCacheClient,
	}
}

func (f ProcessInstanceIDFetcher) Fetch(appGUID string) (map[int]string, error) {
	res, err := f.client.PromQLRange(
		context.Background(),
		fmt.Sprintf(`absolute_usage{source_id="%s"}`, appGUID),
		logcache.WithPromQLStart(time.Now().Add(-1*time.Minute)),
		logcache.WithPromQLEnd(time.Now()),
		logcache.WithPromQLStep("1m"),
	)
	if err != nil {
		return nil, err
	}

	processInstanceIDs := map[int]string{}
	latestFirstPointTimes := map[int]float64{}

	for _, series := range res.GetMatrix().GetSeries() {
		metric := series.GetMetric()
		instanceID, err := strconv.Atoi(metric["instance_id"])
		if err != nil {
			continue
		}
		points := series.GetPoints()
		if len(points) == 0 {
			continue
		}
		firstPointTime, err := strconv.ParseFloat(points[0].GetTime(), 64)
		if err != nil {
			continue
		}
		instanceFirstPointTime, isPresent := latestFirstPointTimes[instanceID]
		if !isPresent || firstPointTime > instanceFirstPointTime {
			latestFirstPointTimes[instanceID] = firstPointTime
			processInstanceIDs[instanceID] = metric["process_instance_id"]
		}
	}
	return processInstanceIDs, nil
}
