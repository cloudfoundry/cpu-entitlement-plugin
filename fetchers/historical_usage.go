package fetchers

import (
	"context"
	"fmt"
	"time"

	"strconv"

	"code.cloudfoundry.org/cpu-entitlement-plugin/metadata"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	logcache "code.cloudfoundry.org/log-cache/pkg/client"
	"code.cloudfoundry.org/log-cache/pkg/rpc/logcache_v1"
)

const QueryStep = 15 * time.Second

type InstanceData struct {
	Time       time.Time
	InstanceID int
	Value      float64
}

type HistoricalUsageFetcher struct {
	client LogCacheClient
	from   time.Time
	to     time.Time
}

//go:generate counterfeiter . LogCacheClient

type LogCacheClient interface {
	Read(ctx context.Context, sourceID string, start time.Time, opts ...logcache.ReadOption) ([]*loggregator_v2.Envelope, error)
	PromQL(ctx context.Context, query string, opts ...logcache.PromQLOption) (*logcache_v1.PromQL_InstantQueryResult, error)
	PromQLRange(ctx context.Context, query string, opts ...logcache.PromQLOption) (*logcache_v1.PromQL_RangeQueryResult, error)
}

func NewHistoricalUsageFetcher(client LogCacheClient, from, to time.Time) *HistoricalUsageFetcher {
	return &HistoricalUsageFetcher{client: client, from: from, to: to}
}

func (f HistoricalUsageFetcher) FetchInstanceData(appGUID string, appInstances map[int]metadata.CFAppInstance) (map[int][]InstanceData, error) {
	res, err := f.client.PromQLRange(
		context.Background(),
		fmt.Sprintf(`absolute_usage{source_id="%s"} / absolute_entitlement{source_id="%s"}`, appGUID, appGUID),
		logcache.WithPromQLStart(f.from),
		logcache.WithPromQLEnd(f.to),
		logcache.WithPromQLStep(QueryStep.String()),
	)
	if err != nil {
		return nil, err
	}

	return parseResult(res, appInstances), nil
}

func parseResult(res *logcache_v1.PromQL_RangeQueryResult, appInstances map[int]metadata.CFAppInstance) map[int][]InstanceData {
	dataPerInstance := map[int][]InstanceData{}
	for _, series := range res.GetMatrix().GetSeries() {
		instanceID, err := strconv.Atoi(series.GetMetric()["instance_id"])
		if err != nil {
			continue
		}

		instance, instanceExists := appInstances[instanceID]
		if !instanceExists {
			continue
		}

		if !isCurrentSeries(series, instance) {
			continue
		}

		var instanceDataRow []InstanceData
		for _, point := range series.GetPoints() {
			timestamp, err := strconv.ParseFloat(point.GetTime(), 64)
			if err != nil {
				continue
			}
			instanceData := InstanceData{
				InstanceID: instanceID,
				Time:       time.Unix(int64(timestamp), 0),
				Value:      point.Value,
			}
			instanceDataRow = append(instanceDataRow, instanceData)
		}

		dataPerInstance[instanceID] = instanceDataRow
	}

	return dataPerInstance
}

func isCurrentSeries(series *logcache_v1.PromQL_Series, instance metadata.CFAppInstance) bool {
	points := series.GetPoints()
	if len(points) == 0 {
		return false
	}

	timestamp, err := strconv.ParseFloat(points[0].GetTime(), 64)
	if err != nil {
		return false
	}
	pointTime := time.Unix(int64(timestamp), 0)
	return pointTime.After(instance.Since) || pointTime.Equal(instance.Since)
}
