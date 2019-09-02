package fetchers

import (
	"context"
	"fmt"
)

type CumulativeUsageFetcher struct {
	logCacheClient LogCacheClient
}

func NewCumulativeUsageFetcher(logCacheClient LogCacheClient) CumulativeUsageFetcher {
	return CumulativeUsageFetcher{logCacheClient: logCacheClient}
}

func (f CumulativeUsageFetcher) FetchInstanceEntitlementUsages(appGuid string) ([]float64, error) {
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
