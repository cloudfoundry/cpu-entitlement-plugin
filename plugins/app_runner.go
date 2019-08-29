package plugins

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cpu-entitlement-plugin/cf"
	"code.cloudfoundry.org/cpu-entitlement-plugin/reporter"
	"code.cloudfoundry.org/cpu-entitlement-plugin/result"
	"github.com/fatih/color"
)

//go:generate counterfeiter . CFClient

type CFClient interface {
	GetApplication(appName string) (cf.Application, error)
}

//go:generate counterfeiter . OutputRenderer

type OutputRenderer interface {
	ShowInstanceReports(cf.Application, []reporter.InstanceReport) error
	ShowMessage(cf.Application, string, ...interface{})
}

//go:generate counterfeiter . Reporter

type Reporter interface {
	CreateInstanceReports(app cf.Application) ([]reporter.InstanceReport, error)
}

type AppRunner struct {
	cfClient        CFClient
	reporter        Reporter
	metricsRenderer OutputRenderer
}

func NewAppRunner(cfClient CFClient, reporter Reporter, metricsRenderer OutputRenderer) AppRunner {
	return AppRunner{
		cfClient:        cfClient,
		reporter:        reporter,
		metricsRenderer: metricsRenderer,
	}
}

func (r AppRunner) Run(appName string) result.Result {
	info, err := r.cfClient.GetApplication(appName)
	if err != nil {
		return result.FailureFromError(err)
	}

	if len(info.Instances) == 0 {
		r.metricsRenderer.ShowMessage(info, "There are no running instances of this process.")
		return result.Success()
	}

	instanceReports, err := r.reporter.CreateInstanceReports(info)
	if err != nil {
		return result.FailureFromError(err).WithWarning(bold("Your Cloud Foundry may not have enabled the CPU Entitlements feature. Please consult your operator."))
	}

	if len(instanceReports) == 0 {
		return result.Failure(fmt.Sprintf("Could not find any CPU data for app %s. Make sure that you are using cf-deployment version >= v5.5.0.", appName))
	}

	err = r.metricsRenderer.ShowInstanceReports(info, instanceReports)
	if err != nil {
		return result.FailureFromError(err)
	}

	return result.Success()
}

func bold(message string) string {
	return terminal.Colorize(message, color.Bold)
}
