package calculator

import (
	"sort"
	"time"

	"code.cloudfoundry.org/cpu-entitlement-plugin/metrics"
)

type Calculator struct{}

type InstanceReport struct {
	InstanceId       int
	EntitlementUsage float64
	LastSpikeFrom    time.Time
	LastSpikeTo      time.Time

	isComplete bool
}

func (r InstanceReport) hasRecordedSpike() bool {
	return !r.LastSpikeTo.IsZero()
}

func (r InstanceReport) hasNotRecordedSpikeYet() bool {
	return !r.hasRecordedSpike()
}

func New() Calculator {
	return Calculator{}
}

func (c Calculator) CalculateInstanceReports(instancesData []metrics.InstanceData) []InstanceReport {
	latestReports := map[int]*InstanceReport{}

	for _, instanceData := range instancesData {
		report, exists := latestReports[instanceData.InstanceId]
		if !exists {
			report = calculateInstanceReport(instanceData)
			latestReports[instanceData.InstanceId] = report
		}

		if report.isComplete {
			continue
		}

		if isSpiking(instanceData) {
			if report.hasNotRecordedSpikeYet() {
				report.LastSpikeTo = instanceData.Time
			}
			report.LastSpikeFrom = instanceData.Time
		}

		if !isSpiking(instanceData) && report.hasRecordedSpike() {
			report.isComplete = true
		}
	}

	return buildReportsSlice(latestReports)
}

func calculateInstanceReport(data metrics.InstanceData) *InstanceReport {
	return &InstanceReport{
		InstanceId:       data.InstanceId,
		EntitlementUsage: entitlementUsage(data),
	}
}

func isSpiking(instanceData metrics.InstanceData) bool {
	return entitlementUsage(instanceData) > 1
}

func entitlementUsage(data metrics.InstanceData) float64 {
	return data.AbsoluteUsage / data.AbsoluteEntitlement
}

func buildReportsSlice(reportsMap map[int]*InstanceReport) []InstanceReport {
	var reports []InstanceReport
	for _, report := range reportsMap {
		reports = append(reports, *report)
	}

	sort.Slice(reports, func(i, j int) bool {
		return reports[i].InstanceId < reports[j].InstanceId
	})

	return reports
}
