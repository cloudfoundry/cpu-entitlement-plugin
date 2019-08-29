package plugins

import (
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cpu-entitlement-plugin/reporter"
	"code.cloudfoundry.org/cpu-entitlement-plugin/result"
	"github.com/fatih/color"
)

//go:generate counterfeiter . OutputRenderer

type OutputRenderer interface {
	ShowApplicationReport(appReport reporter.ApplicationReport) error
}

//go:generate counterfeiter . Reporter

type Reporter interface {
	CreateApplicationReport(appName string) (reporter.ApplicationReport, error)
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

func (r AppRunner) Run(appName string) result.Result {
	applicationReport, err := r.reporter.CreateApplicationReport(appName)
	if err != nil {
		if _, ok := err.(reporter.UnsupportedCFDeploymentError); ok {
			return result.FailureFromError(err)
		}

		return result.FailureFromError(err).WithWarning(bold("Your Cloud Foundry may not have enabled the CPU Entitlements feature. Please consult your operator."))
	}

	err = r.metricsRenderer.ShowApplicationReport(applicationReport)
	if err != nil {
		return result.FailureFromError(err)
	}

	return result.Success()
}

func bold(message string) string {
	return terminal.Colorize(message, color.Bold)
}
