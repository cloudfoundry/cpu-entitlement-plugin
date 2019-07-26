package plugin

import (
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cpu-entitlement-plugin/calculator"
	"code.cloudfoundry.org/cpu-entitlement-plugin/metadata"
	"code.cloudfoundry.org/cpu-entitlement-plugin/metrics"
	"code.cloudfoundry.org/cpu-entitlement-plugin/result"
	"github.com/fatih/color"
)

//go:generate counterfeiter . CFAppInfoGetter

type CFAppInfoGetter interface {
	GetCFAppInfo(appName string) (metadata.CFAppInfo, error)
}

//go:generate counterfeiter . MetricsFetcher

type MetricsFetcher interface {
	FetchAll(appGUID string, instanceCount int) ([]metrics.InstanceData, error)
}

//go:generate counterfeiter . MetricsRenderer

type MetricsRenderer interface {
	ShowInstanceReports(metadata.CFAppInfo, []calculator.InstanceReport) error
}

//go:generate counterfeiter . MetricsCalculator

type MetricsCalculator interface {
	CalculateInstanceReports(usages []metrics.InstanceData) []calculator.InstanceReport
}

type Runner struct {
	infoGetter        CFAppInfoGetter
	metricsFetcher    MetricsFetcher
	metricsCalculator MetricsCalculator
	metricsRenderer   MetricsRenderer
}

func NewRunner(infoGetter CFAppInfoGetter, metricsFetcher MetricsFetcher, metricsCalculator MetricsCalculator, metricsRenderer MetricsRenderer) Runner {
	return Runner{
		infoGetter:        infoGetter,
		metricsFetcher:    metricsFetcher,
		metricsCalculator: metricsCalculator,
		metricsRenderer:   metricsRenderer,
	}
}

func (r Runner) Run(appName string) result.Result {
	info, err := r.infoGetter.GetCFAppInfo(appName)
	if err != nil {
		return result.FailureFromError(err)
	}

	usageMetrics, err := r.metricsFetcher.FetchAll(info.App.Guid, info.App.InstanceCount)
	if err != nil {
		return result.FailureFromError(err).WithWarning(bold("Your Cloud Foundry may not have enabled the CPU Entitlements feature. Please consult your operator."))
	}

	instanceReports := r.metricsCalculator.CalculateInstanceReports(usageMetrics)

	err = r.metricsRenderer.ShowInstanceReports(info, instanceReports)
	if err != nil {
		return result.FailureFromError(err)
	}

	return result.Success()
}

func bold(message string) string {
	return terminal.Colorize(message, color.Bold)
}
