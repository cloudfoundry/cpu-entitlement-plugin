package fetchers

import (
	"context"
	"fmt"
	"strconv"

	"code.cloudfoundry.org/cpu-entitlement-plugin/cf"
	"code.cloudfoundry.org/log-cache/pkg/rpc/logcache_v1"
)

//go:generate counterfeiter . Fetcher
type Fetcher interface {
	FetchInstanceData(appGUID string, appInstances map[int]cf.Instance) (map[int]InstanceData, error)
}

type CurrentUsageFetcher struct {
	client          LogCacheClient
	fallbackFetcher Fetcher
}

func NewCurrentUsageFetcher(client LogCacheClient) CurrentUsageFetcher {
	return CurrentUsageFetcher{
		client:          client,
		fallbackFetcher: NewCumulativeUsageFetcher(client),
	}
}

func NewCurrentUsageFetcherWithFallbackFetcher(client LogCacheClient, fallbackFetcher Fetcher) CurrentUsageFetcher {
	return CurrentUsageFetcher{
		client:          client,
		fallbackFetcher: fallbackFetcher,
	}
}

func (f CurrentUsageFetcher) FetchInstanceData(appGUID string, appInstances map[int]cf.Instance) (map[int]InstanceData, error) {
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

	fallbackResult, err := f.fallbackFetcher.FetchInstanceData(appGUID, appInstances)
	if err != nil {
		return nil, err
	}

	for instanceID, _ := range appInstances {
		if _, has := deltaResult[instanceID]; !has {
			if result, ok := fallbackResult[instanceID]; ok {
				deltaResult[instanceID] = result
				continue
			}
			return nil, fmt.Errorf("could not find historical usage for instance ID %d", instanceID)
		}
	}

	return deltaResult, nil
}

func parseCurrentUsage(res *logcache_v1.PromQL_InstantQueryResult, appInstances map[int]cf.Instance) map[int]InstanceData {
	usagePerInstance := map[int]InstanceData{}
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
		usagePerInstance[instanceID] = dataPoint
	}

	return usagePerInstance
}
