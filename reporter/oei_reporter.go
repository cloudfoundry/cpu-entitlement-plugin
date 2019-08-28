package reporter

import "code.cloudfoundry.org/cpu-entitlement-plugin/cf"

type OEIReport struct {
	SpaceReports []SpaceReport
}

type SpaceReport struct {
	SpaceName string
	Apps      []string
}

//go:generate counterfeiter . MetricsFetcher

type MetricsFetcher interface {
	FetchInstanceEntitlementUsages(appGuid string) ([]float64, error)
}

//go:generate counterfeiter . CloudFoundryClient

type CloudFoundryClient interface {
	GetSpaces() ([]cf.Space, error)
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
	spaceReports := []SpaceReport{}

	spaces, err := r.cf.GetSpaces()
	if err != nil {
		return OEIReport{}, err
	}

	for _, space := range spaces {
		apps, err := r.filterApps(space.Applications)
		if err != nil {
			return OEIReport{}, err
		}

		if len(apps) == 0 {
			continue
		}

		spaceReports = append(spaceReports, SpaceReport{SpaceName: space.Name, Apps: apps})
	}

	return OEIReport{SpaceReports: spaceReports}, nil
}

func (r OverEntitlementInstances) filterApps(spaceApps []cf.Application) ([]string, error) {
	apps := []string{}
	for _, app := range spaceApps {
		isOverEntitlement, err := r.isOverEntitlement(app.Guid)
		if err != nil {
			return nil, err
		}
		if isOverEntitlement {
			apps = append(apps, app.Name)
		}
	}
	return apps, nil
}

func (r OverEntitlementInstances) isOverEntitlement(appGuid string) (bool, error) {
	appInstancesUsages, err := r.metricsFetcher.FetchInstanceEntitlementUsages(appGuid)
	if err != nil {
		return false, err
	}

	isOverEntitlement := false
	for _, usage := range appInstancesUsages {
		if usage > 1 {
			isOverEntitlement = true
		}
	}

	return isOverEntitlement, nil
}
