package calculator

import (
	"sort"

	"code.cloudfoundry.org/cpu-entitlement-plugin/metrics"
)

type Calculator struct{}

type InstanceReport struct {
	InstanceId       int
	EntitlementUsage float64
}

func New() Calculator {
	return Calculator{}
}

func (c Calculator) CalculateInstanceReports(instancesData []metrics.InstanceData) []InstanceReport {
	latestReports := map[int]InstanceReport{}
	for _, instanceData := range instancesData {
		_, exists := latestReports[instanceData.InstanceId]
		if !exists {
			latestReports[instanceData.InstanceId] = c.calculateInstanceReport(instanceData)
		}
	}

	return buildReportsSlice(latestReports)
}

func (c Calculator) calculateInstanceReport(data metrics.InstanceData) InstanceReport {
	return InstanceReport{
		InstanceId:       data.InstanceId,
		EntitlementUsage: data.AbsoluteUsage / data.AbsoluteEntitlement,
	}
}

func buildReportsSlice(reportsMap map[int]InstanceReport) []InstanceReport {
	var reports []InstanceReport
	for _, report := range reportsMap {
		reports = append(reports, report)
	}

	sort.Slice(reports, func(i, j int) bool {
		return reports[i].InstanceId < reports[j].InstanceId
	})

	return reports
}
