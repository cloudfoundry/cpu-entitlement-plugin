package plugins

import (
	"code.cloudfoundry.org/cpu-entitlement-plugin/reporter"
	"code.cloudfoundry.org/lager"
)

//go:generate counterfeiter . OverEntitlementInstancesReporter

type OverEntitlementInstancesReporter interface {
	OverEntitlementInstances(logger lager.Logger) (reporter.OEIReport, error)
}

//go:generate counterfeiter . OverEntitlementInstancesRenderer

type OverEntitlementInstancesRenderer interface {
	Render(lager.Logger, reporter.OEIReport) error
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

func (r *OverEntitlementInstancesRunner) Run(logger lager.Logger) error {
	logger = logger.Session("run")
	logger.Info("start")
	defer logger.Info("end")

	report, err := r.reporter.OverEntitlementInstances(logger)
	if err != nil {
		return err
	}

	err = r.renderer.Render(logger, report)
	if err != nil {
		return err
	}

	return nil
}
