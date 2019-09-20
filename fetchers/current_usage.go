package fetchers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"code.cloudfoundry.org/cpu-entitlement-plugin/cf"
	"code.cloudfoundry.org/log-cache/pkg/rpc/logcache_v1"
)

//go:generate counterfeiter . HistoricalFetcher
type HistoricalFetcher interface {
	FetchInstanceData(appGUID string, appInstances map[int]cf.Instance) (map[int][]InstanceData, error)
}

type CurrentUsageFetcher struct {
	client            LogCacheClient
	historicalFetcher HistoricalFetcher
}

func NewCurrentUsageFetcher(client LogCacheClient, from, to time.Time) CurrentUsageFetcher {
	return CurrentUsageFetcher{
		client:            client,
		historicalFetcher: NewHistoricalUsageFetcher(client, from, to),
	}
}

func NewCurrentUsageFetcherWithHistoricalFetcher(client LogCacheClient, historicalFetcher HistoricalFetcher) CurrentUsageFetcher {
	return CurrentUsageFetcher{
		client:            client,
		historicalFetcher: historicalFetcher,
	}
}

func (f CurrentUsageFetcher) FetchInstanceData(appGUID string, appInstances map[int]cf.Instance) (map[int][]InstanceData, error) {
	res, err := f.client.PromQL(
		context.Background(),
		fmt.Sprintf(`idelta(absolute_usage{source_id="%s"}[1m]) / idelta(absolute_entitlement{source_id="%s"}[1m])`, appGUID, appGUID),
	)
	if err != nil {
		return nil, err
	}

	deltaResult := parseCurrentUsage(res, appInstances)
	if len(deltaResult) == len(appInstances) {
		return deltaResult, nil
	}

	historicalResult, err := f.historicalFetcher.FetchInstanceData(appGUID, appInstances)
	if err != nil {
		return nil, err
	}

	for instanceID, _ := range appInstances {
		if _, has := deltaResult[instanceID]; !has {
			if result, ok := historicalResult[instanceID]; ok {
				deltaResult[instanceID] = []InstanceData{result[len(result)-1]}
				continue
			}
			return nil, fmt.Errorf("could not find historical usage for instance ID %d", instanceID)
		}
	}

	return deltaResult, nil
}

func parseCurrentUsage(res *logcache_v1.PromQL_InstantQueryResult, appInstances map[int]cf.Instance) map[int][]InstanceData {
	usagePerInstance := map[int][]InstanceData{}
	for _, sample := range res.GetVector().GetSamples() {
		instanceID, err := strconv.Atoi(sample.GetMetric()["instance_id"])
		if err != nil {
			continue
		}

		processInstanceID := sample.GetMetric()["process_instance_id"]
		if processInstanceID != appInstances[instanceID].ProcessInstanceID {
			continue
		}

		dataPoint := InstanceData{
			InstanceID: instanceID,
			Value:      sample.GetPoint().GetValue(),
		}
		usagePerInstance[instanceID] = []InstanceData{dataPoint}
	}

	return usagePerInstance
}
