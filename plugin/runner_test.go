package plugin_test

import (
	"errors"

	"code.cloudfoundry.org/cpu-entitlement-plugin/metadata"
	"code.cloudfoundry.org/cpu-entitlement-plugin/plugin"
	"code.cloudfoundry.org/cpu-entitlement-plugin/plugin/pluginfakes"
	"code.cloudfoundry.org/cpu-entitlement-plugin/reporter"
	"code.cloudfoundry.org/cpu-entitlement-plugin/result"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Runner", func() {
	var (
		infoGetter       *pluginfakes.FakeCFAppInfoGetter
		instanceReporter *pluginfakes.FakeReporter
		metricsRenderer  *pluginfakes.FakeMetricsRenderer

		runner    plugin.Runner
		runResult result.Result
	)

	BeforeEach(func() {
		infoGetter = new(pluginfakes.FakeCFAppInfoGetter)
		instanceReporter = new(pluginfakes.FakeReporter)
		metricsRenderer = new(pluginfakes.FakeMetricsRenderer)
		runner = plugin.NewRunner(infoGetter, instanceReporter, metricsRenderer)

		infoGetter.GetCFAppInfoReturns(metadata.CFAppInfo{
			Guid: "123",
			Name: "app-name",
		}, nil)

		instanceReporter.CreateInstanceReportsReturns([]reporter.InstanceReport{
			{
				InstanceID: 0,
				HistoricalUsage: reporter.HistoricalUsage{
					Value: 0.5,
				},
			},
			{
				InstanceID: 1,
				HistoricalUsage: reporter.HistoricalUsage{
					Value: 0.8,
				},
			},
			{
				InstanceID: 2,
				HistoricalUsage: reporter.HistoricalUsage{
					Value: 0.875,
				},
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

		Expect(instanceReporter.CreateInstanceReportsCallCount()).To(Equal(1))
		Expect(instanceReporter.CreateInstanceReportsArgsForCall(0)).To(Equal("123"))

		Expect(metricsRenderer.ShowInstanceReportsCallCount()).To(Equal(1))
		info, instanceReports := metricsRenderer.ShowInstanceReportsArgsForCall(0)
		Expect(info).To(Equal(metadata.CFAppInfo{
			Guid: "123",
			Name: "app-name",
		}))
		Expect(instanceReports).To(Equal([]reporter.InstanceReport{
			{
				InstanceID: 0,
				HistoricalUsage: reporter.HistoricalUsage{
					Value: 0.5,
				},
			},
			{
				InstanceID: 1,
				HistoricalUsage: reporter.HistoricalUsage{
					Value: 0.8,
				},
			},
			{
				InstanceID: 2,
				HistoricalUsage: reporter.HistoricalUsage{
					Value: 0.875,
				},
			},
		}))
	})

	When("getting the app info fails", func() {
		BeforeEach(func() {
			infoGetter.GetCFAppInfoReturns(metadata.CFAppInfo{}, errors.New("info error"))
		})

		It("returns a failure with a warning", func() {
			Expect(runResult.IsFailure).To(BeTrue())
			Expect(runResult.ErrorMessage).To(Equal("info error"))
		})
	})

	When("creating the reports fails", func() {
		BeforeEach(func() {
			instanceReporter.CreateInstanceReportsReturns(nil, errors.New("reports error"))
		})

		It("returns a failure", func() {
			Expect(runResult.IsFailure).To(BeTrue())
			Expect(runResult.ErrorMessage).To(Equal("reports error"))
			Expect(runResult.WarningMessage).To(ContainSubstring("Your Cloud Foundry may not have enabled the CPU Entitlements feature. Please consult your operator."))
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
