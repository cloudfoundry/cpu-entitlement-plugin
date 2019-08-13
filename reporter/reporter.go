package reporter

import (
	"fmt"
	"time"

	"code.cloudfoundry.org/cpu-entitlement-plugin/fetchers"
	"code.cloudfoundry.org/cpu-entitlement-plugin/metadata"
)

type Reporter struct {
	currentUsageFetcher InstanceDataFetcher
	averageUsageFetcher InstanceDataFetcher
	lastSpikeFetcher    InstanceDataFetcher
}

//go:generate counterfeiter . InstanceDataFetcher

type InstanceDataFetcher interface {
	FetchInstanceData(appGUID string, appInstances map[int]metadata.CFAppInstance) (map[int][]fetchers.InstanceData, error)
}

type InstanceReport struct {
	InstanceID      int
	HistoricalUsage HistoricalUsage
	CurrentUsage    CurrentUsage
}

type HistoricalUsage struct {
	Value         float64
	LastSpikeFrom time.Time
	LastSpikeTo   time.Time
}

type CurrentUsage struct {
	Value float64
}

func (r InstanceReport) HasRecordedSpike() bool {
	return !r.HistoricalUsage.LastSpikeTo.IsZero()
}

func New(currentUsageFetcher, averageUsageFetcher, lastSpikeFetcher InstanceDataFetcher) Reporter {
	return Reporter{
		currentUsageFetcher: currentUsageFetcher,
		averageUsageFetcher: averageUsageFetcher,
		lastSpikeFetcher:    lastSpikeFetcher,
	}
}

func (r Reporter) CreateInstanceReports(appInfo metadata.CFAppInfo) ([]InstanceReport, error) {
	latestReports := map[int]InstanceReport{}

	currentUsagePerInstance, err := r.currentUsageFetcher.FetchInstanceData(appInfo.Guid, appInfo.Instances)
	if err != nil {
		return nil, err
	}

	averageUsagePerInstance, err := r.averageUsageFetcher.FetchInstanceData(appInfo.Guid, appInfo.Instances)
	if err != nil {
		return nil, err
	}

	lastSpikePerInstance, err := r.lastSpikeFetcher.FetchInstanceData(appInfo.Guid, appInfo.Instances)
	if err != nil {
		return nil, err
	}

	for _, instance := range appInfo.Instances {
		if len(currentUsagePerInstance[instance.InstanceID]) == 0 {
			return nil, fmt.Errorf("no current usage for instance id %d", instance.InstanceID)
		}
		currentUsage := currentUsagePerInstance[instance.InstanceID][0]

		if len(averageUsagePerInstance[instance.InstanceID]) == 0 {
			return nil, fmt.Errorf("no average usage for instance id %d", instance.InstanceID)
		}
		averageUsage := averageUsagePerInstance[instance.InstanceID][0]

		report := InstanceReport{
			InstanceID:      instance.InstanceID,
			CurrentUsage:    CurrentUsage{Value: currentUsage.Value},
			HistoricalUsage: HistoricalUsage{Value: averageUsage.Value},
		}

		if len(lastSpikePerInstance[instance.InstanceID]) > 0 {
			lastSpike := lastSpikePerInstance[instance.InstanceID][0]
			report.HistoricalUsage.LastSpikeFrom = lastSpike.From
			report.HistoricalUsage.LastSpikeTo = lastSpike.To
		}

		latestReports[instance.InstanceID] = report
	}

	return sortByInstance(latestReports), nil
}

func sortByInstance(reports map[int]InstanceReport) []InstanceReport {
	result := []InstanceReport{}
	for i := 0; i < len(reports); i++ {
		result = append(result, reports[i])
	}

	return result
}

// Instance #0 was over entitlement from 2019-08-13 15:32:33 to 2019-08-13 15:34:03
