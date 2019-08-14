package plugin

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cpu-entitlement-plugin/metadata"
	"code.cloudfoundry.org/cpu-entitlement-plugin/reporter"
	"code.cloudfoundry.org/cpu-entitlement-plugin/result"
	"github.com/fatih/color"
)

//go:generate counterfeiter . CFAppInfoGetter

type CFAppInfoGetter interface {
	GetCFAppInfo(appName string) (metadata.CFAppInfo, error)
}

//go:generate counterfeiter . MetricsRenderer

type MetricsRenderer interface {
	ShowInstanceReports(metadata.CFAppInfo, []reporter.InstanceReport) error
}

//go:generate counterfeiter . Reporter

type Reporter interface {
	CreateInstanceReports(appInfo metadata.CFAppInfo) ([]reporter.InstanceReport, error)
}

type Runner struct {
	infoGetter      CFAppInfoGetter
	reporter        Reporter
	metricsRenderer MetricsRenderer
}

func NewRunner(infoGetter CFAppInfoGetter, reporter Reporter, metricsRenderer MetricsRenderer) Runner {
	return Runner{
		infoGetter:      infoGetter,
		reporter:        reporter,
		metricsRenderer: metricsRenderer,
	}
}

func (r Runner) Run(appName string) result.Result {
	info, err := r.infoGetter.GetCFAppInfo(appName)
	if err != nil {
		return result.FailureFromError(err)
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
