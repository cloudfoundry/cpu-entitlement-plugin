package plugin

import (
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cpu-entitlement-plugin/metadata"
	"code.cloudfoundry.org/cpu-entitlement-plugin/result"
	"code.cloudfoundry.org/cpu-entitlement-plugin/usagemetric"
	"github.com/fatih/color"
)

//go:generate counterfeiter . CFAppInfoGetter

type CFAppInfoGetter interface {
	GetCFAppInfo(appName string) (metadata.CFAppInfo, error)
}

//go:generate counterfeiter . MetricFetcher

type MetricFetcher interface {
	FetchLatest(appGUID string, instanceCount int) ([]usagemetric.UsageMetric, error)
}

//go:generate counterfeiter . MetricsRenderer

type MetricsRenderer interface {
	ShowMetrics(metadata.CFAppInfo, []usagemetric.UsageMetric) error
}

type Runner struct {
	infoGetter      CFAppInfoGetter
	metricFetcher   MetricFetcher
	metricsRenderer MetricsRenderer
}

func NewRunner(infoGetter CFAppInfoGetter, metricFetcher MetricFetcher, metricsRenderer MetricsRenderer) Runner {
	return Runner{infoGetter: infoGetter, metricFetcher: metricFetcher, metricsRenderer: metricsRenderer}
}

func (r Runner) Run(appName string) result.Result {
	info, err := r.infoGetter.GetCFAppInfo(appName)
	if err != nil {
		return result.FailureFromError(err)
	}

	usageMetrics, err := r.metricFetcher.FetchLatest(info.App.Guid, info.App.InstanceCount)
	if err != nil {
		return result.FailureFromError(err).WithWarning(bold("Your Cloud Foundry may not have enabled the CPU Entitlements feature. Please consult your operator."))
	}

	err = r.metricsRenderer.ShowMetrics(info, usageMetrics)
	if err != nil {
		return result.FailureFromError(err)
	}

	return result.Success()
}

func bold(message string) string {
	return terminal.Colorize(message, color.Bold)
}
