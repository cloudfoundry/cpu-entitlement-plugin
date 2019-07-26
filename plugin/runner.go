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

//go:generate counterfeiter . MetricFetcher

type MetricFetcher interface {
	FetchLatest(appGUID string, instanceCount int) ([]metrics.Usage, error)
}

//go:generate counterfeiter . MetricsRenderer

type MetricsRenderer interface {
	ShowInfos(metadata.CFAppInfo, []calculator.InstanceInfo) error
}

//go:generate counterfeiter . MetricsCalculator

type MetricsCalculator interface {
	CalculateInstanceInfos(usages []metrics.Usage) []calculator.InstanceInfo
}

type Runner struct {
	infoGetter        CFAppInfoGetter
	metricFetcher     MetricFetcher
	metricsCalculator MetricsCalculator
	metricsRenderer   MetricsRenderer
}

func NewRunner(infoGetter CFAppInfoGetter, metricFetcher MetricFetcher, metricsCalculator MetricsCalculator, metricsRenderer MetricsRenderer) Runner {
	return Runner{
		infoGetter:        infoGetter,
		metricFetcher:     metricFetcher,
		metricsCalculator: metricsCalculator,
		metricsRenderer:   metricsRenderer,
	}
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

	instanceInfos := r.metricsCalculator.CalculateInstanceInfos(usageMetrics)

	err = r.metricsRenderer.ShowInfos(info, instanceInfos)
	if err != nil {
		return result.FailureFromError(err)
	}

	return result.Success()
}

func bold(message string) string {
	return terminal.Colorize(message, color.Bold)
}
