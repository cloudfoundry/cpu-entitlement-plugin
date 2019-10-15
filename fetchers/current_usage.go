package fetchers

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"code.cloudfoundry.org/cpu-entitlement-plugin/cf"
	"code.cloudfoundry.org/log-cache/pkg/rpc/logcache_v1"
)

type CurrentInstanceData struct {
	InstanceID int
	Usage      float64
}

//go:generate counterfeiter . Fetcher
type Fetcher interface {
	FetchInstanceData(appGUID string, appInstances map[int]cf.Instance) (map[int]interface{}, error)
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

func (f CurrentUsageFetcher) FetchInstanceData(appGUID string, appInstances map[int]cf.Instance) (map[int]interface{}, error) {
	res, err := f.client.PromQL(
		context.Background(),
		fmt.Sprintf(`idelta(absolute_usage{source_id="%s"}[1m]) / idelta(absolute_entitlement{source_id="%s"}[1m])`, appGUID, appGUID),
	)
	if err != nil {
		return nil, err
	}

	currentUsage := parseCurrentUsage(res, appInstances)
	if len(currentUsage) == len(appInstances) {
		return currentUsage, nil
	}

	cumulativeResult, err := f.fallbackFetcher.FetchInstanceData(appGUID, appInstances)
	if err != nil {
		return nil, err
	}

	for instanceID, _ := range appInstances {
		if _, ok := currentUsage[instanceID]; ok {
			continue
		}

		result, ok := cumulativeResult[instanceID]
		if !ok {
			return nil, fmt.Errorf("could not find historical usage for instance ID %d", instanceID)
		}

		cumulativeData, ok := result.(CumulativeInstanceData)
		if !ok {
			return map[int]interface{}{}, errors.New("")
		}

		currentUsage[instanceID] = CurrentInstanceData{
			InstanceID: cumulativeData.InstanceID,
			Usage:      cumulativeData.Usage,
		}
	}

	return currentUsage, nil
}

func parseCurrentUsage(res *logcache_v1.PromQL_InstantQueryResult, appInstances map[int]cf.Instance) map[int]interface{} {
	usagePerInstance := make(map[int]interface{})
	for _, sample := range res.GetVector().GetSamples() {
		instanceID, err := strconv.Atoi(sample.GetMetric()["instance_id"])
		if err != nil {
			continue
		}

		processInstanceID := sample.GetMetric()["process_instance_id"]
		if processInstanceID != appInstances[instanceID].ProcessInstanceID {
			continue
		}

		dataPoint := CurrentInstanceData{
			InstanceID: instanceID,
			Usage:      sample.GetPoint().GetValue(),
		}
		usagePerInstance[instanceID] = dataPoint
	}

	return usagePerInstance
}
