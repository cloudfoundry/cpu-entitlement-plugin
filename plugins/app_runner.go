package plugins

import (
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cpu-entitlement-plugin/reporter"
	"code.cloudfoundry.org/cpu-entitlement-plugin/result"
	"code.cloudfoundry.org/lager"
	"github.com/fatih/color"
)

//go:generate counterfeiter . OutputRenderer

type OutputRenderer interface {
	ShowApplicationReport(appReport reporter.ApplicationReport) error
}

//go:generate counterfeiter . Reporter

type Reporter interface {
	CreateApplicationReport(logger lager.Logger, appName string) (reporter.ApplicationReport, error)
}

type AppRunner struct {
	reporter        Reporter
	metricsRenderer OutputRenderer
}

func NewAppRunner(reporter Reporter, metricsRenderer OutputRenderer) AppRunner {
	return AppRunner{
		reporter:        reporter,
		metricsRenderer: metricsRenderer,
	}
}

func (r AppRunner) Run(logger lager.Logger, appName string) result.Result {
	logger = logger.Session("run", lager.Data{"app-name": appName})
	logger.Info("start")
	defer logger.Info("end")
	applicationReport, err := r.reporter.CreateApplicationReport(logger, appName)
	if err != nil {
		if _, ok := err.(reporter.UnsupportedCFDeploymentError); ok {
			logger.Error("unsupported-cf-deployment", err)
			return result.FailureFromError(err)
		}

		logger.Error("failed-creating-app-report", err)
		return result.FailureFromError(err).WithWarning(bold("Your Cloud Foundry may not have enabled the CPU Entitlements feature. Please consult your operator."))
	}

	err = r.metricsRenderer.ShowApplicationReport(applicationReport)
	if err != nil {
		logger.Error("failed-rendering-app-metrics", err)
		return result.FailureFromError(err)
	}

	return result.Success()
}

func bold(message string) string {
	return terminal.Colorize(message, color.Bold)
}
