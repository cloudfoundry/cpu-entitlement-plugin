package reporter // import "code.cloudfoundry.org/cpu-entitlement-admin-plugin/reporter"

type Report struct {
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
	GetSpaces() ([]Space, error)
}

type Space struct {
	Name         string
	Applications []Application
}

type Application struct {
	Name string
	Guid string
}

type Reporter struct {
	cf             CloudFoundryClient
	metricsFetcher MetricsFetcher
}

func New(cf CloudFoundryClient, metricsFetcher MetricsFetcher) Reporter {
	return Reporter{
		cf:             cf,
		metricsFetcher: metricsFetcher,
	}
}

func (r Reporter) OverEntitlementInstances() (Report, error) {
	spaceReports := []SpaceReport{}

	spaces, err := r.cf.GetSpaces()
	if err != nil {
		return Report{}, err
	}

	for _, space := range spaces {
		apps, err := r.filterApps(space.Applications)
		if err != nil {
			return Report{}, err
		}

		if len(apps) == 0 {
			continue
		}

		spaceReports = append(spaceReports, SpaceReport{SpaceName: space.Name, Apps: apps})
	}

	return Report{SpaceReports: spaceReports}, nil
}

func (r Reporter) filterApps(spaceApps []Application) ([]string, error) {
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

func (r Reporter) isOverEntitlement(appGuid string) (bool, error) {
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
