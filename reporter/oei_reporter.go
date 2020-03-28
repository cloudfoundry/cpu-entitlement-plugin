package reporter

import (
	"sort"

	"code.cloudfoundry.org/cpu-entitlement-plugin/cf"
	"code.cloudfoundry.org/cpu-entitlement-plugin/fetchers"
	"code.cloudfoundry.org/lager"
)

type OEIReport struct {
	Org          string
	Username     string
	SpaceReports []SpaceReport
}

type SpaceReport struct {
	SpaceName string
	Apps      []string
}

//go:generate counterfeiter . MetricsFetcher

type MetricsFetcher interface {
	FetchInstanceData(logger lager.Logger, appGuid string, appInstances map[int]cf.Instance) (map[int]interface{}, error)
}

//go:generate counterfeiter . CloudFoundryClient

type CloudFoundryClient interface {
	GetSpaces() ([]cf.Space, error)
	GetCurrentOrg() (string, error)
	Username() (string, error)
}

type OverEntitlementInstances struct {
	cf             CloudFoundryClient
	metricsFetcher MetricsFetcher
}

func NewOverEntitlementInstances(cf CloudFoundryClient, metricsFetcher MetricsFetcher) OverEntitlementInstances {
	return OverEntitlementInstances{
		cf:             cf,
		metricsFetcher: metricsFetcher,
	}
}

func (r OverEntitlementInstances) OverEntitlementInstances(logger lager.Logger) (OEIReport, error) {
	logger = logger.Session("oei-reporter")
	logger.Info("start")
	defer logger.Info("end")

	org, err := r.cf.GetCurrentOrg()
	if err != nil {
		logger.Error("failed-to-get-current-org", err)
		return OEIReport{}, err
	}

	user, err := r.cf.Username()
	if err != nil {
		logger.Error("failed-to-get-username", err)
		return OEIReport{}, err
	}

	spaces, err := r.cf.GetSpaces()
	if err != nil {
		logger.Error("failed-to-get-spaces", err)
		return OEIReport{}, err
	}

	spaceReports, err := r.buildSpaceReports(logger, spaces)
	if err != nil {
		return OEIReport{}, err
	}

	return OEIReport{Org: org, Username: user, SpaceReports: spaceReports}, nil
}

func (r OverEntitlementInstances) buildSpaceReports(logger lager.Logger, spaces []cf.Space) ([]SpaceReport, error) {
	spaceReports := []SpaceReport{}
	for _, space := range spaces {
		apps, err := r.filterApps(logger, space.Applications)
		if err != nil {
			return nil, err
		}

		if len(apps) == 0 {
			continue
		}
		sort.Strings(apps)
		spaceReports = append(spaceReports, SpaceReport{SpaceName: space.Name, Apps: apps})
	}

	sort.Slice(spaceReports, func(i, j int) bool {
		return spaceReports[i].SpaceName < spaceReports[j].SpaceName
	})

	return spaceReports, nil
}

func (r OverEntitlementInstances) filterApps(logger lager.Logger, spaceApps []cf.Application) ([]string, error) {
	apps := []string{}
	for _, app := range spaceApps {
		isOverEntitlement, err := r.isOverEntitlement(logger, app.Guid, app.Instances)
		if err != nil {
			return nil, err
		}
		if isOverEntitlement {
			apps = append(apps, app.Name)
		}
	}
	return apps, nil
}

func (r OverEntitlementInstances) isOverEntitlement(logger lager.Logger, appGuid string, appInstances map[int]cf.Instance) (bool, error) {
	logger = logger.Session("is-over-entitlement", lager.Data{"app-guid": appGuid})
	appInstancesUsages, err := r.metricsFetcher.FetchInstanceData(logger, appGuid, appInstances)
	if err != nil {
		logger.Error("failed-to-fetch-instance-metrics", err)
		return false, err
	}

	isOverEntitlement := false
	for _, instanceData := range appInstancesUsages {
		cumulativeInstanceData, ok := instanceData.(fetchers.CumulativeInstanceData)
		if !ok {
			logger.Info("metrics-fetcher-returned-wrong-type",
				lager.Data{"instance-data": instanceData})
			continue
		}

		if cumulativeInstanceData.Usage > 1 {
			isOverEntitlement = true
		}
	}

	return isOverEntitlement, nil
}
