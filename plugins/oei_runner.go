package plugins

import (
	"code.cloudfoundry.org/cpu-entitlement-plugin/reporter"
	"code.cloudfoundry.org/lager"
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

func (r *OverEntitlementInstancesRunner) Run(logger lager.Logger) error {
	logger = logger.Session("run")
	logger.Info("start")
	defer logger.Info("end")

	report, err := r.reporter.OverEntitlementInstances()
	if err != nil {
		logger.Error("failed-creating-oei-report", err)
		return err
	}

	err = r.renderer.Render(report)
	if err != nil {
		logger.Error("failed-rendering-oei-metrics", err)
		return err
	}

	return nil
}
