package fetchers

import (
	"context"
	"fmt"
	"time"

	"strconv"

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

func (f HistoricalUsageFetcher) FetchInstanceData(appGUID string) (map[int][]InstanceData, error) {
	procInstanceIDs, err := f.getActiveProcInstanceIDs(appGUID)
	if err != nil {
		return nil, err
	}

	return f.fetchPeriod(appGUID, procInstanceIDs)
}

func (f HistoricalUsageFetcher) getActiveProcInstanceIDs(appGUID string) (map[string]bool, error) {
	appsSnapshot, err := f.client.PromQL(context.Background(), fmt.Sprintf(`absolute_usage{source_id="%s"}`, appGUID))
	if err != nil {
		return nil, err
	}

	procInstanceIDs := map[string]bool{}
	for _, sample := range appsSnapshot.GetVector().GetSamples() {
		processInstanceID := sample.GetMetric()["process_instance_id"]
		procInstanceIDs[processInstanceID] = true
	}

	return procInstanceIDs, nil
}

func (f HistoricalUsageFetcher) fetchPeriod(appGUID string, procInstanceIDs map[string]bool) (map[int][]InstanceData, error) {
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

	return parseResult(res, procInstanceIDs), nil
}

func parseResult(res *logcache_v1.PromQL_RangeQueryResult, procInstanceIDs map[string]bool) map[int][]InstanceData {
	dataPerInstance := map[int][]InstanceData{}
	for _, series := range res.GetMatrix().GetSeries() {
		if !procInstanceIDs[series.GetMetric()["process_instance_id"]] {
			continue
		}

		instanceID, err := strconv.Atoi(series.GetMetric()["instance_id"])
		if err != nil {
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
