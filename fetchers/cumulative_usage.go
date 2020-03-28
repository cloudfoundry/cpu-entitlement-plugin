package fetchers

import (
	"context"
	"fmt"
	"strconv"

	"code.cloudfoundry.org/cpu-entitlement-plugin/cf"
	"code.cloudfoundry.org/lager"
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

func (f CumulativeUsageFetcher) FetchInstanceData(logger lager.Logger, appGuid string, appInstances map[int]cf.Instance) (map[int]interface{}, error) {
	logger = logger.Session("cumulative-usage-fetcher", lager.Data{"app-guid": appGuid})
	logger.Info("start")
	defer logger.Info("end")

	query := fmt.Sprintf(`absolute_usage{source_id="%s"} / absolute_entitlement{source_id="%s"}`, appGuid, appGuid)
	promqlResult, err := f.logCacheClient.PromQL(context.Background(), query)
	if err != nil {
		logger.Error("promql-failed", err, lager.Data{"query": query})
		return nil, err
	}

	instanceUsages := make(map[int]interface{})
	for _, sample := range promqlResult.GetVector().GetSamples() {
		instanceID, err := strconv.Atoi(sample.GetMetric()["instance_id"])
		if err != nil {
			logger.Info("ignoring-corrupt-instance-id", lager.Data{"instance-id": sample.GetMetric()["instance_id"]})
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
