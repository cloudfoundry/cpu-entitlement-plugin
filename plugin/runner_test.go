package plugin_test

import (
	"errors"

	models "code.cloudfoundry.org/cli/plugin/models"
	"code.cloudfoundry.org/cpu-entitlement-plugin/calculator"
	"code.cloudfoundry.org/cpu-entitlement-plugin/metadata"
	"code.cloudfoundry.org/cpu-entitlement-plugin/metrics"
	"code.cloudfoundry.org/cpu-entitlement-plugin/plugin"
	"code.cloudfoundry.org/cpu-entitlement-plugin/plugin/pluginfakes"
	"code.cloudfoundry.org/cpu-entitlement-plugin/result"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Runner", func() {
	var (
		infoGetter        *pluginfakes.FakeCFAppInfoGetter
		metricsFetcher    *pluginfakes.FakeMetricsFetcher
		metricsCalculator *pluginfakes.FakeMetricsCalculator
		metricsRenderer   *pluginfakes.FakeMetricsRenderer

		runner    plugin.Runner
		runResult result.Result
	)

	BeforeEach(func() {
		infoGetter = new(pluginfakes.FakeCFAppInfoGetter)
		metricsFetcher = new(pluginfakes.FakeMetricsFetcher)
		metricsCalculator = new(pluginfakes.FakeMetricsCalculator)
		metricsRenderer = new(pluginfakes.FakeMetricsRenderer)
		runner = plugin.NewRunner(infoGetter, metricsFetcher, metricsCalculator, metricsRenderer)

		infoGetter.GetCFAppInfoReturns(metadata.CFAppInfo{
			App: models.GetAppModel{
				Guid:          "123",
				Name:          "app-name",
				InstanceCount: 3,
			},
		}, nil)

		metricsFetcher.FetchAllReturns([]metrics.InstanceData{
			{
				InstanceId:          0,
				AbsoluteUsage:       1.0,
				AbsoluteEntitlement: 2.0,
				ContainerAge:        3.0,
			},
			{
				InstanceId:          2,
				AbsoluteUsage:       7.0,
				AbsoluteEntitlement: 8.0,
				ContainerAge:        9.0,
			},
			{
				InstanceId:          1,
				AbsoluteUsage:       4.0,
				AbsoluteEntitlement: 5.0,
				ContainerAge:        6.0,
			},
		}, nil)

		metricsCalculator.CalculateInstanceReportsReturns([]calculator.InstanceReport{
			{
				InstanceId:       0,
				EntitlementUsage: 0.5,
			},
			{
				InstanceId:       1,
				EntitlementUsage: 0.8,
			},
			{
				InstanceId:       2,
				EntitlementUsage: 0.875,
			},
		})
	})

	JustBeforeEach(func() {
		runResult = runner.Run("app-name")
	})

	It("prints the app CPU metrics", func() {
		Expect(runResult.IsFailure).To(BeFalse())

		Expect(infoGetter.GetCFAppInfoCallCount()).To(Equal(1))
		appName := infoGetter.GetCFAppInfoArgsForCall(0)
		Expect(appName).To(Equal("app-name"))

		Expect(metricsFetcher.FetchAllCallCount()).To(Equal(1))
		guid, instanceCount := metricsFetcher.FetchAllArgsForCall(0)
		Expect(guid).To(Equal("123"))
		Expect(instanceCount).To(Equal(3))

		Expect(metricsCalculator.CalculateInstanceReportsCallCount()).To(Equal(1))
		usageMetrics := metricsCalculator.CalculateInstanceReportsArgsForCall(0)
		Expect(usageMetrics).To(Equal([]metrics.InstanceData{
			{
				InstanceId:          0,
				AbsoluteUsage:       1.0,
				AbsoluteEntitlement: 2.0,
				ContainerAge:        3.0,
			},
			{
				InstanceId:          2,
				AbsoluteUsage:       7.0,
				AbsoluteEntitlement: 8.0,
				ContainerAge:        9.0,
			},
			{
				InstanceId:          1,
				AbsoluteUsage:       4.0,
				AbsoluteEntitlement: 5.0,
				ContainerAge:        6.0,
			},
		}))

		Expect(metricsRenderer.ShowInstanceReportsCallCount()).To(Equal(1))
		info, instanceReports := metricsRenderer.ShowInstanceReportsArgsForCall(0)
		Expect(info).To(Equal(metadata.CFAppInfo{
			App: models.GetAppModel{
				Guid:          "123",
				Name:          "app-name",
				InstanceCount: 3,
			},
		}))
		Expect(instanceReports).To(Equal([]calculator.InstanceReport{
			{
				InstanceId:       0,
				EntitlementUsage: 0.5,
			},
			{
				InstanceId:       1,
				EntitlementUsage: 0.8,
			},
			{
				InstanceId:       2,
				EntitlementUsage: 0.875,
			},
		}))
	})

	When("getting the app info fails", func() {
		BeforeEach(func() {
			infoGetter.GetCFAppInfoReturns(metadata.CFAppInfo{}, errors.New("info error"))
		})

		It("returns a failure", func() {
			Expect(runResult.IsFailure).To(BeTrue())
			Expect(runResult.ErrorMessage).To(Equal("info error"))
		})
	})

	When("fetching the app metrics fails", func() {
		BeforeEach(func() {
			metricsFetcher.FetchAllReturns(nil, errors.New("metrics error"))
		})

		It("returns a failure", func() {
			Expect(runResult.IsFailure).To(BeTrue())
			Expect(runResult.ErrorMessage).To(Equal("metrics error"))
		})
	})

	When("rendering the app metrics fails", func() {
		BeforeEach(func() {
			metricsRenderer.ShowInstanceReportsReturns(errors.New("render error"))
		})

		It("returns a failure", func() {
			Expect(runResult.IsFailure).To(BeTrue())
			Expect(runResult.ErrorMessage).To(Equal("render error"))
		})
	})
})
