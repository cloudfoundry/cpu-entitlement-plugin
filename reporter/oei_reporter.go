package reporter

import (
	"sort"

	"code.cloudfoundry.org/cpu-entitlement-plugin/cf"
	"code.cloudfoundry.org/cpu-entitlement-plugin/fetchers"
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
	FetchInstanceData(appGuid string, appInstances map[int]cf.Instance) (map[int]fetchers.InstanceData, error)
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

func (r OverEntitlementInstances) OverEntitlementInstances() (OEIReport, error) {
	org, err := r.cf.GetCurrentOrg()
	if err != nil {
		return OEIReport{}, err
	}

	user, err := r.cf.Username()
	if err != nil {
		return OEIReport{}, err
	}

	spaces, err := r.cf.GetSpaces()
	if err != nil {
		return OEIReport{}, err
	}

	spaceReports, err := r.buildSpaceReports(spaces)
	if err != nil {
		return OEIReport{}, err
	}

	return OEIReport{Org: org, Username: user, SpaceReports: spaceReports}, nil
}

func (r OverEntitlementInstances) buildSpaceReports(spaces []cf.Space) ([]SpaceReport, error) {
	spaceReports := []SpaceReport{}
	for _, space := range spaces {
		apps, err := r.filterApps(space.Applications)
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

func (r OverEntitlementInstances) filterApps(spaceApps []cf.Application) ([]string, error) {
	apps := []string{}
	for _, app := range spaceApps {
		isOverEntitlement, err := r.isOverEntitlement(app.Guid, app.Instances)
		if err != nil {
			return nil, err
		}
		if isOverEntitlement {
			apps = append(apps, app.Name)
		}
	}
	return apps, nil
}

func (r OverEntitlementInstances) isOverEntitlement(appGuid string, appInstances map[int]cf.Instance) (bool, error) {
	appInstancesUsages, err := r.metricsFetcher.FetchInstanceData(appGuid, appInstances)
	if err != nil {
		return false, err
	}

	isOverEntitlement := false
	for _, usage := range appInstancesUsages {
		if usage.Value > 1 {
			isOverEntitlement = true
		}
	}

	return isOverEntitlement, nil
}
