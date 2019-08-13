package fetchers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"code.cloudfoundry.org/cpu-entitlement-plugin/metadata"
	"code.cloudfoundry.org/log-cache/pkg/rpc/logcache_v1"
)

type AverageUsageFetcher struct {
	client LogCacheClient
}

func NewAverageUsageFetcher(client LogCacheClient) *AverageUsageFetcher {
	return &AverageUsageFetcher{client: client}
}

func (f AverageUsageFetcher) FetchInstanceData(appGUID string, appInstances map[int]metadata.CFAppInstance) (map[int][]InstanceData, error) {
	res, err := f.client.PromQL(
		context.Background(),
		fmt.Sprintf(`absolute_usage{source_id="%s"} / absolute_entitlement{source_id="%s"}`, appGUID, appGUID),
	)
	if err != nil {
		return nil, err
	}

	return parseAverageUsage(res, appInstances), nil
}

func parseAverageUsage(res *logcache_v1.PromQL_InstantQueryResult, appInstances map[int]metadata.CFAppInstance) map[int][]InstanceData {
	usagePerInstance := map[int][]InstanceData{}
	for _, sample := range res.GetVector().GetSamples() {
		instanceID, err := strconv.Atoi(sample.GetMetric()["instance_id"])
		if err != nil {
			continue
		}
		timestamp, err := strconv.ParseFloat(sample.GetPoint().GetTime(), 64)
		if err != nil {
			continue
		}

		dataPoint := InstanceData{
			InstanceID: instanceID,
			Time:       time.Unix(int64(timestamp), 0),
			Value:      sample.GetPoint().GetValue(),
		}
		if isValid(dataPoint, appInstances) {
			usagePerInstance[instanceID] = []InstanceData{dataPoint}
		}
	}

	return usagePerInstance
}
