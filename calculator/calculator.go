package calculator

import (
	"sort"
	"time"

	"code.cloudfoundry.org/cpu-entitlement-plugin/metrics"
)

type Calculator struct{}

type InstanceReport struct {
	InstanceID       int
	EntitlementUsage float64
	LastSpikeFrom    time.Time
	LastSpikeTo      time.Time
}

func (r InstanceReport) HasRecordedSpike() bool {
	return !r.LastSpikeTo.IsZero()
}

func New() Calculator {
	return Calculator{}
}

func (c Calculator) CalculateInstanceReports(dataPerInstance map[int][]metrics.InstanceData) []InstanceReport {
	latestReports := map[int]InstanceReport{}

	for instanceID, instanceData := range dataPerInstance {
		spikeFrom, spikeTo := findLatestSpike(instanceData)
		latestReports[instanceID] = InstanceReport{
			InstanceID:       instanceID,
			EntitlementUsage: instanceData[len(instanceData)-1].EntitlementUsage,
			LastSpikeFrom:    spikeFrom,
			LastSpikeTo:      spikeTo,
		}
	}

	return buildReportsSlice(latestReports)
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
