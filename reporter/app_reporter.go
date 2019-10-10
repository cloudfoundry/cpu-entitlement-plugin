package reporter

import (
	"fmt"
	"sort"
	"time"

	"code.cloudfoundry.org/cpu-entitlement-plugin/cf"
	"code.cloudfoundry.org/cpu-entitlement-plugin/fetchers"
)

type UnsupportedCFDeploymentError struct {
	message string
}

func (e UnsupportedCFDeploymentError) Error() string {
	return e.message
}

func NewUnsupportedCFDeploymentError(appName string) error {
	return UnsupportedCFDeploymentError{message: fmt.Sprintf("Could not find any CPU data for app %s. Make sure that you are using cf-deployment version >= v5.5.0.", appName)}
}

type AppReporter struct {
	currentUsageFetcher    InstanceDataFetcher
	lastSpikeFetcher       InstanceDataFetcher
	cumulativeUsageFetcher InstanceDataFetcher
	cfClient               AppReporterCloudFoundryClient
}

//go:generate counterfeiter . InstanceDataFetcher

type InstanceDataFetcher interface {
	FetchInstanceData(appGUID string, appInstances map[int]cf.Instance) (map[int][]fetchers.InstanceData, error)
}

//go:generate counterfeiter . AppReporterCloudFoundryClient
type AppReporterCloudFoundryClient interface {
	GetApplication(appName string) (cf.Application, error)
	GetCurrentOrg() (string, error)
	GetCurrentSpace() (string, error)
	Username() (string, error)
}

type ApplicationReport struct {
	Org             string
	Username        string
	Space           string
	ApplicationName string
	InstanceReports []InstanceReport
}

type InstanceReport struct {
	InstanceID      int
	CumulativeUsage CumulativeUsage
	CurrentUsage    CurrentUsage
	LastSpike       LastSpike
}

type LastSpike struct {
	From time.Time
	To   time.Time
}

type CurrentUsage struct {
	Value float64
}

type CumulativeUsage struct {
	Value float64
}

func NewAppReporter(cfClient AppReporterCloudFoundryClient, currentUsageFetcher, lastSpikeFetcher, cumulativeUsageFetcher InstanceDataFetcher) AppReporter {
	return AppReporter{
		cfClient:               cfClient,
		currentUsageFetcher:    currentUsageFetcher,
		lastSpikeFetcher:       lastSpikeFetcher,
		cumulativeUsageFetcher: cumulativeUsageFetcher,
	}
}

func (r AppReporter) CreateApplicationReport(appName string) (ApplicationReport, error) {
	application, err := r.cfClient.GetApplication(appName)
	if err != nil {
		return ApplicationReport{}, err
	}

	org, err := r.cfClient.GetCurrentOrg()
	if err != nil {
		return ApplicationReport{}, err
	}

	space, err := r.cfClient.GetCurrentSpace()
	if err != nil {
		return ApplicationReport{}, err
	}

	user, err := r.cfClient.Username()
	if err != nil {
		return ApplicationReport{}, err
	}

	if len(application.Instances) == 0 {
		return ApplicationReport{Org: org, Space: space, Username: user, ApplicationName: appName}, nil
	}

	latestReports := map[int]InstanceReport{}

	currentUsagePerInstance, err := r.currentUsageFetcher.FetchInstanceData(application.Guid, application.Instances)
	if err != nil {
		return ApplicationReport{}, err
	}
	if len(currentUsagePerInstance) == 0 {
		return ApplicationReport{}, NewUnsupportedCFDeploymentError(appName)
	}

	for instanceID, currentUsage := range currentUsagePerInstance {
		if len(currentUsage) != 1 {
			continue
		}
		currentReport := getOrCreateInstanceReport(latestReports, instanceID)
		currentReport.CurrentUsage = CurrentUsage{Value: currentUsage[0].Value}
		latestReports[instanceID] = currentReport
	}

	lastSpikePerInstance, err := r.lastSpikeFetcher.FetchInstanceData(application.Guid, application.Instances)
	if err != nil {
		return ApplicationReport{}, err
	}

	for instanceID, lastSpike := range lastSpikePerInstance {
		currentReport := getOrCreateInstanceReport(latestReports, instanceID)
		currentReport.LastSpike = LastSpike{
			From: lastSpike[0].From,
			To:   lastSpike[0].To,
		}
		latestReports[instanceID] = currentReport
	}

	cumulativeUsagePerInstance, err := r.cumulativeUsageFetcher.FetchInstanceData(application.Guid, application.Instances)
	if err != nil {
		return ApplicationReport{}, err
	}

	for instanceID, cumulativeUsage := range cumulativeUsagePerInstance {
		currentReport := getOrCreateInstanceReport(latestReports, instanceID)
		currentReport.CumulativeUsage = CumulativeUsage{
			Value: cumulativeUsage[0].Value,
		}
		latestReports[instanceID] = currentReport
	}
	instanceReports := buildReportsSlice(latestReports)

	return ApplicationReport{Org: org, Space: space, Username: user, ApplicationName: appName, InstanceReports: instanceReports}, nil
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
