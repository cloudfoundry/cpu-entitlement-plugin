package plugin_test

import (
	"errors"
	"time"

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

		metricsFetcher.FetchInstanceDataReturns(map[int][]metrics.InstanceData{
			0: {
				{
					Time:             time.Unix(1, 0),
					InstanceID:       0,
					EntitlementUsage: 0.5,
				},
			},
			2: {
				{
					Time:             time.Unix(2, 0),
					InstanceID:       2,
					EntitlementUsage: 0.6,
				},
			},
			1: {
				{
					Time:             time.Unix(3, 0),
					InstanceID:       1,
					EntitlementUsage: 0.7,
				},
			},
		}, nil)

		metricsCalculator.CalculateInstanceReportsReturns([]calculator.InstanceReport{
			{
				InstanceID:       0,
				EntitlementUsage: 0.5,
			},
			{
				InstanceID:       1,
				EntitlementUsage: 0.8,
			},
			{
				InstanceID:       2,
				EntitlementUsage: 0.875,
			},
		})
	})

	JustBeforeEach(func() {
		runResult = runner.Run("app-name", time.Unix(123, 0), time.Unix(456, 0))
	})

	It("prints the app CPU metrics", func() {
		Expect(runResult.IsFailure).To(BeFalse())

		Expect(infoGetter.GetCFAppInfoCallCount()).To(Equal(1))
		appName := infoGetter.GetCFAppInfoArgsForCall(0)
		Expect(appName).To(Equal("app-name"))

		Expect(metricsFetcher.FetchInstanceDataCallCount()).To(Equal(1))
		guid, from, to := metricsFetcher.FetchInstanceDataArgsForCall(0)
		Expect(guid).To(Equal("123"))
		Expect(from).To(Equal(time.Unix(123, 0)))
		Expect(to).To(Equal(time.Unix(456, 0)))

		Expect(metricsCalculator.CalculateInstanceReportsCallCount()).To(Equal(1))
		usageMetrics := metricsCalculator.CalculateInstanceReportsArgsForCall(0)
		Expect(usageMetrics).To(Equal(map[int][]metrics.InstanceData{
			0: {
				{
					Time:             time.Unix(1, 0),
					InstanceID:       0,
					EntitlementUsage: 0.5,
				},
			},
			2: {
				{
					Time:             time.Unix(2, 0),
					InstanceID:       2,
					EntitlementUsage: 0.6,
				},
			},
			1: {
				{
					Time:             time.Unix(3, 0),
					InstanceID:       1,
					EntitlementUsage: 0.7,
				},
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
				InstanceID:       0,
				EntitlementUsage: 0.5,
			},
			{
				InstanceID:       1,
				EntitlementUsage: 0.8,
			},
			{
				InstanceID:       2,
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
			metricsFetcher.FetchInstanceDataReturns(nil, errors.New("metrics error"))
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
