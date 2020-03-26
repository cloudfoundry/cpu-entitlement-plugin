package reporter

import (
	"fmt"
	"sort"
	"time"

	"code.cloudfoundry.org/cpu-entitlement-plugin/cf"
	"code.cloudfoundry.org/cpu-entitlement-plugin/fetchers"
	"code.cloudfoundry.org/lager"
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
	FetchInstanceData(appGUID string, appInstances map[int]cf.Instance) (map[int]interface{}, error)
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

func (r AppReporter) CreateApplicationReport(logger lager.Logger, appName string) (ApplicationReport, error) {
	logger = logger.Session("create-application-report", lager.Data{"app-name": appName})
	logger.Info("start")
	defer logger.Info("end")

	application, err := r.cfClient.GetApplication(appName)
	if err != nil {
		logger.Error("failed-to-get-app", err)
		return ApplicationReport{}, err
	}

	org, err := r.cfClient.GetCurrentOrg()
	if err != nil {
		logger.Error("failed-to-get-current-org", err)
		return ApplicationReport{}, err
	}

	space, err := r.cfClient.GetCurrentSpace()
	if err != nil {
		logger.Error("failed-to-get-current-space", err)
		return ApplicationReport{}, err
	}

	user, err := r.cfClient.Username()
	if err != nil {
		logger.Error("failed-to-get-username", err)
		return ApplicationReport{}, err
	}

	if len(application.Instances) == 0 {
		logger.Info("no-instances-found-for-app")
		return ApplicationReport{Org: org, Space: space, Username: user, ApplicationName: appName}, nil
	}

	latestReports := map[int]InstanceReport{}

	currentUsagePerInstance, err := r.currentUsageFetcher.FetchInstanceData(application.Guid, application.Instances)
	if err != nil {
		logger.Error("failed-to-fetch-current-usage", err)
		return ApplicationReport{}, err
	}
	if len(currentUsagePerInstance) == 0 {
		err = NewUnsupportedCFDeploymentError(appName)
		logger.Error("no-current-usage-data-found", err)
		return ApplicationReport{}, err
	}

	for instanceID, instanceData := range currentUsagePerInstance {
		currentInstanceData, ok := instanceData.(fetchers.CurrentInstanceData)
		if !ok {
			logger.Info("current-usage-reporter-returned-wrong-type",
				lager.Data{"instance-data": instanceData})
			continue
		}

		currentReport := getOrCreateInstanceReport(latestReports, instanceID)
		currentReport.CurrentUsage = CurrentUsage{Value: currentInstanceData.Usage}
		latestReports[instanceID] = currentReport
	}

	lastSpikePerInstance, err := r.lastSpikeFetcher.FetchInstanceData(application.Guid, application.Instances)
	if err != nil {
		logger.Error("failed-to-fetch-last-spikes", err)
		return ApplicationReport{}, err
	}

	for instanceID, data := range lastSpikePerInstance {
		lastSpikeInstanceData, ok := data.(fetchers.LastSpikeInstanceData)
		if !ok {
			logger.Info("last-spike-reporter-returned-wrong-type",
				lager.Data{"instance-data": data})
			continue
		}
		currentReport := getOrCreateInstanceReport(latestReports, instanceID)
		currentReport.LastSpike = LastSpike{
			From: lastSpikeInstanceData.From,
			To:   lastSpikeInstanceData.To,
		}
		latestReports[instanceID] = currentReport
	}

	cumulativeUsagePerInstance, err := r.cumulativeUsageFetcher.FetchInstanceData(application.Guid, application.Instances)
	if err != nil {
		logger.Error("failed-to-fetch-cumulative-usage", err)
		return ApplicationReport{}, err
	}

	for instanceID, data := range cumulativeUsagePerInstance {
		cumulativeUsageInstanceData, ok := data.(fetchers.CumulativeInstanceData)
		if !ok {
			logger.Info("cumulative-usage-reporter-returned-wrong-type",
				lager.Data{"instance-data": data})
			continue
		}
		currentReport := getOrCreateInstanceReport(latestReports, instanceID)
		currentReport.CumulativeUsage = CumulativeUsage{
			Value: cumulativeUsageInstanceData.Usage,
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
