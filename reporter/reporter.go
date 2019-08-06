package reporter

import (
	"sort"
	"time"

	"code.cloudfoundry.org/cpu-entitlement-plugin/metrics"
)

type Reporter struct {
	metricsFetcher MetricsFetcher
}

//go:generate counterfeiter . MetricsFetcher

type MetricsFetcher interface {
	FetchInstanceData(appGUID string, from, to time.Time) (map[int][]metrics.InstanceData, error)
}

type InstanceReport struct {
	InstanceID       int
	EntitlementUsage float64
	LastSpikeFrom    time.Time
	LastSpikeTo      time.Time
}

func (r InstanceReport) HasRecordedSpike() bool {
	return !r.LastSpikeTo.IsZero()
}

func New(metricsFetcher MetricsFetcher) Reporter {
	return Reporter{
		metricsFetcher: metricsFetcher,
	}
}

func (r Reporter) CreateInstanceReports(appGUID string, from, to time.Time) ([]InstanceReport, error) {
	latestReports := map[int]InstanceReport{}

	dataPerInstance, err := r.metricsFetcher.FetchInstanceData(appGUID, from, to)
	if err != nil {
		return nil, err
	}

	for instanceID, instanceData := range dataPerInstance {
		spikeFrom, spikeTo := findLatestSpike(instanceData)
		latestReports[instanceID] = InstanceReport{
			InstanceID:       instanceID,
			EntitlementUsage: instanceData[len(instanceData)-1].EntitlementUsage,
			LastSpikeFrom:    spikeFrom,
			LastSpikeTo:      spikeTo,
		}
	}

	return buildReportsSlice(latestReports), nil
}

func findLatestSpike(instanceData []metrics.InstanceData) (time.Time, time.Time) {
	var from, to time.Time

	for i := len(instanceData) - 1; i >= 0; i-- {
		dataPoint := instanceData[i]

		if isSpiking(dataPoint) {
			if to.IsZero() {
				to = dataPoint.Time
			}
			from = dataPoint.Time
		}

		if !isSpiking(dataPoint) && !to.IsZero() {
			break
		}
	}

	return from, to
}

func isSpiking(dataPoint metrics.InstanceData) bool {
	return dataPoint.EntitlementUsage > 1
}

func buildReportsSlice(reportsMap map[int]InstanceReport) []InstanceReport {
	var reports []InstanceReport
	for _, report := range reportsMap {
		reports = append(reports, report)
	}

	sort.Slice(reports, func(i, j int) bool {
		return reports[i].InstanceID < reports[j].InstanceID
	})

	return reports
}
