package fetchers

import (
	"context"
	"fmt"
	"strconv"

	"code.cloudfoundry.org/cpu-entitlement-plugin/cf"
)

type CumulativeInstanceData struct {
	InstanceID int
	Usage      float64
}

type CumulativeUsageFetcher struct {
	logCacheClient LogCacheClient
}

func NewCumulativeUsageFetcher(logCacheClient LogCacheClient) CumulativeUsageFetcher {
	return CumulativeUsageFetcher{logCacheClient: logCacheClient}
}

func (f CumulativeUsageFetcher) FetchInstanceData(appGuid string, appInstances map[int]cf.Instance) (map[int]interface{}, error) {
	promqlResult, err := f.logCacheClient.PromQL(context.Background(),
		fmt.Sprintf(`absolute_usage{source_id="%s"} / absolute_entitlement{source_id="%s"}`, appGuid, appGuid))
	if err != nil {
		return nil, err
	}

	instanceUsages := make(map[int]interface{})
	for _, sample := range promqlResult.GetVector().GetSamples() {
		instanceID, err := strconv.Atoi(sample.GetMetric()["instance_id"])
		if err != nil {
			continue
		}
		processInstanceID := sample.GetMetric()["process_instance_id"]
		if appInstances[instanceID].ProcessInstanceID != processInstanceID {
			continue
		}

		instanceUsages[instanceID] = CumulativeInstanceData{
			InstanceID: instanceID,
			Usage:      sample.GetPoint().GetValue(),
		}
	}

	return instanceUsages, nil
}
