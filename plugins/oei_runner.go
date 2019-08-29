package plugins

import (
	"code.cloudfoundry.org/cpu-entitlement-plugin/reporter"
)

//go:generate counterfeiter . OverEntitlementInstancesReporter

type OverEntitlementInstancesReporter interface {
	OverEntitlementInstances() (reporter.OEIReport, error)
}

//go:generate counterfeiter . OverEntitlementInstancesRenderer

type OverEntitlementInstancesRenderer interface {
	Render(reporter.OEIReport) error
}

type OverEntitlementInstancesRunner struct {
	reporter OverEntitlementInstancesReporter
	renderer OverEntitlementInstancesRenderer
}

func NewOverEntitlementInstancesRunner(oeiReporter OverEntitlementInstancesReporter, oeiRenderer OverEntitlementInstancesRenderer) *OverEntitlementInstancesRunner {
	return &OverEntitlementInstancesRunner{
		reporter: oeiReporter,
		renderer: oeiRenderer,
	}
}

func (r *OverEntitlementInstancesRunner) Run() error {
	report, err := r.reporter.OverEntitlementInstances()
	if err != nil {
		return err
	}

	return r.renderer.Render(report)
}
