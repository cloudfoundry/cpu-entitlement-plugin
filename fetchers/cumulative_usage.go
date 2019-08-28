package fetchers

import (
	"context"
	"fmt"
)

type CumulativeUsage struct {
	logCacheClient LogCacheClient
}

func NewCumulativeUsage(logCacheClient LogCacheClient) CumulativeUsage {
	return CumulativeUsage{logCacheClient: logCacheClient}
}

func (f CumulativeUsage) FetchInstanceEntitlementUsages(appGuid string) ([]float64, error) {
	promqlResult, err := f.logCacheClient.PromQL(context.Background(),
		fmt.Sprintf(`absolute_usage{source_id="%s"} / absolute_entitlement{source_id="%s"}`, appGuid, appGuid))
	if err != nil {
		return nil, err
	}

	var instanceUsages []float64
	for _, sample := range promqlResult.GetVector().GetSamples() {
		instanceUsages = append(instanceUsages, sample.GetPoint().GetValue())
	}

	return instanceUsages, nil
}
