package fetchers

import (
	"context"
	"fmt"
	"strconv"

	"code.cloudfoundry.org/cpu-entitlement-plugin/cf"
)

type CumulativeUsageFetcher struct {
	logCacheClient LogCacheClient
}

func NewCumulativeUsageFetcher(logCacheClient LogCacheClient) CumulativeUsageFetcher {
	return CumulativeUsageFetcher{logCacheClient: logCacheClient}
}

func (f CumulativeUsageFetcher) FetchInstanceEntitlementUsages(appGuid string, appInstances map[int]cf.Instance) ([]float64, error) {
	promqlResult, err := f.logCacheClient.PromQL(context.Background(),
		fmt.Sprintf(`absolute_usage{source_id="%s"} / absolute_entitlement{source_id="%s"}`, appGuid, appGuid))
	if err != nil {
		return nil, err
	}

	var instanceUsages []float64
	for _, sample := range promqlResult.GetVector().GetSamples() {
		instanceID, err := strconv.Atoi(sample.GetMetric()["instance_id"])
		if err != nil {
			continue
		}
		processInstanceID := sample.GetMetric()["process_instance_id"]
		if appInstances[instanceID].ProcessInstanceID != processInstanceID {
			continue
		}
		instanceUsages = append(instanceUsages, sample.GetPoint().GetValue())
	}

	return instanceUsages, nil
}
