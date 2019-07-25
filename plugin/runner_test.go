package plugin_test

import (
	"errors"

	models "code.cloudfoundry.org/cli/plugin/models"
	"code.cloudfoundry.org/cpu-entitlement-plugin/metadata"
	"code.cloudfoundry.org/cpu-entitlement-plugin/plugin"
	"code.cloudfoundry.org/cpu-entitlement-plugin/plugin/pluginfakes"
	"code.cloudfoundry.org/cpu-entitlement-plugin/result"
	"code.cloudfoundry.org/cpu-entitlement-plugin/usagemetric"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Runner", func() {
	var (
		infoGetter      *pluginfakes.FakeCFAppInfoGetter
		metricFetcher   *pluginfakes.FakeMetricFetcher
		metricsRenderer *pluginfakes.FakeMetricsRenderer

		runner    plugin.Runner
		runResult result.Result
	)

	BeforeEach(func() {
		infoGetter = new(pluginfakes.FakeCFAppInfoGetter)
		metricFetcher = new(pluginfakes.FakeMetricFetcher)
		metricsRenderer = new(pluginfakes.FakeMetricsRenderer)
		runner = plugin.NewRunner(infoGetter, metricFetcher, metricsRenderer)

		infoGetter.GetCFAppInfoReturns(metadata.CFAppInfo{
			App: models.GetAppModel{
				Guid:          "123",
				Name:          "app-name",
				InstanceCount: 3,
			},
		}, nil)

		metricFetcher.FetchLatestReturns([]usagemetric.UsageMetric{
			{
				InstanceId:          0,
				AbsoluteUsage:       1.0,
				AbsoluteEntitlement: 2.0,
				ContainerAge:        3.0,
			},
			{
				InstanceId:          1,
				AbsoluteUsage:       4.0,
				AbsoluteEntitlement: 5.0,
				ContainerAge:        6.0,
			},
			{
				InstanceId:          2,
				AbsoluteUsage:       7.0,
				AbsoluteEntitlement: 8.0,
				ContainerAge:        9.0,
			},
		}, nil)
	})

	JustBeforeEach(func() {
		runResult = runner.Run("app-name")
	})

	It("prints the app CPU metrics", func() {
		Expect(runResult.IsFailure).To(BeFalse())

		Expect(infoGetter.GetCFAppInfoCallCount()).To(Equal(1))
		appName := infoGetter.GetCFAppInfoArgsForCall(0)
		Expect(appName).To(Equal("app-name"))

		Expect(metricFetcher.FetchLatestCallCount()).To(Equal(1))
		guid, instanceCount := metricFetcher.FetchLatestArgsForCall(0)
		Expect(guid).To(Equal("123"))
		Expect(instanceCount).To(Equal(3))

		Expect(metricsRenderer.ShowMetricsCallCount()).To(Equal(1))
		info, metrics := metricsRenderer.ShowMetricsArgsForCall(0)
		Expect(info).To(Equal(metadata.CFAppInfo{
			App: models.GetAppModel{
				Guid:          "123",
				Name:          "app-name",
				InstanceCount: 3,
			},
		}))
		Expect(metrics).To(Equal([]usagemetric.UsageMetric{
			{
				InstanceId:          0,
				AbsoluteUsage:       1.0,
				AbsoluteEntitlement: 2.0,
				ContainerAge:        3.0,
			},
			{
				InstanceId:          1,
				AbsoluteUsage:       4.0,
				AbsoluteEntitlement: 5.0,
				ContainerAge:        6.0,
			},
			{
				InstanceId:          2,
				AbsoluteUsage:       7.0,
				AbsoluteEntitlement: 8.0,
				ContainerAge:        9.0,
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
			metricFetcher.FetchLatestReturns(nil, errors.New("metrics error"))
		})

		It("returns a failure", func() {
			Expect(runResult.IsFailure).To(BeTrue())
			Expect(runResult.ErrorMessage).To(Equal("metrics error"))
		})
	})

	When("rendering the app metrics fails", func() {
		BeforeEach(func() {
			metricsRenderer.ShowMetricsReturns(errors.New("render error"))
		})

		It("returns a failure", func() {
			Expect(runResult.IsFailure).To(BeTrue())
			Expect(runResult.ErrorMessage).To(Equal("render error"))
		})
	})
})
