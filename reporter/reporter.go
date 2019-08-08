package reporter

import (
	"sort"
	"time"

	"code.cloudfoundry.org/cpu-entitlement-plugin/fetchers"
	"code.cloudfoundry.org/cpu-entitlement-plugin/metadata"
)

type Reporter struct {
	historicalUsageFetcher InstanceDataFetcher
	currentUsageFetcher    InstanceDataFetcher
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

func New(historicalUsageFetcher, currentUsageFetcher InstanceDataFetcher) Reporter {
	return Reporter{
		historicalUsageFetcher: historicalUsageFetcher,
		currentUsageFetcher:    currentUsageFetcher,
	}
}

func (r Reporter) CreateInstanceReports(appInfo metadata.CFAppInfo) ([]InstanceReport, error) {
	latestReports := map[int]InstanceReport{}

	historicalUsagePerInstance, err := r.historicalUsageFetcher.FetchInstanceData(appInfo.Guid, appInfo.Instances)
	if err != nil {
		return nil, err
	}

	for instanceID, historicalUsage := range historicalUsagePerInstance {
		spikeFrom, spikeTo := findLatestSpike(historicalUsage)
		currentReport := getOrCreateInstanceReport(latestReports, instanceID)
		currentReport.HistoricalUsage = HistoricalUsage{
			Value:         historicalUsage[len(historicalUsage)-1].Value,
			LastSpikeFrom: spikeFrom,
			LastSpikeTo:   spikeTo,
		}
		latestReports[instanceID] = currentReport
	}

	currentUsagePerInstance, err := r.currentUsageFetcher.FetchInstanceData(appInfo.Guid, appInfo.Instances)
	if err != nil {
		return nil, err
	}

	for instanceID, currentUsage := range currentUsagePerInstance {
		if len(currentUsage) != 1 {
			continue
		}
		currentReport := getOrCreateInstanceReport(latestReports, instanceID)
		currentReport.CurrentUsage = CurrentUsage{Value: currentUsage[0].Value}
		latestReports[instanceID] = currentReport
	}

	return buildReportsSlice(latestReports), nil
}

func getOrCreateInstanceReport(reports map[int]InstanceReport, instanceID int) InstanceReport {
	_, ok := reports[instanceID]
	if !ok {
		reports[instanceID] = InstanceReport{InstanceID: instanceID}
	}
	return reports[instanceID]
}

func findLatestSpike(instanceData []fetchers.InstanceData) (time.Time, time.Time) {
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

func isSpiking(dataPoint fetchers.InstanceData) bool {
	return dataPoint.Value > 1
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
