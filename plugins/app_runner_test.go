package plugins_test

import (
	"errors"

	"code.cloudfoundry.org/cpu-entitlement-plugin/plugins"
	"code.cloudfoundry.org/cpu-entitlement-plugin/plugins/pluginsfakes"
	"code.cloudfoundry.org/cpu-entitlement-plugin/reporter"
	"code.cloudfoundry.org/cpu-entitlement-plugin/result"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("App Runner", func() {
	var (
		instanceReporter *pluginsfakes.FakeReporter
		outputRenderer   *pluginsfakes.FakeOutputRenderer

		runner            plugins.AppRunner
		runResult         result.Result
		applicationReport reporter.ApplicationReport
	)

	BeforeEach(func() {
		instanceReporter = new(pluginsfakes.FakeReporter)
		outputRenderer = new(pluginsfakes.FakeOutputRenderer)
		runner = plugins.NewAppRunner(instanceReporter, outputRenderer)
		applicationReport = reporter.ApplicationReport{
			InstanceReports: []reporter.InstanceReport{
				{
					InstanceID: 0,
					CumulativeUsage: reporter.CumulativeUsage{
						Value: 0.5,
					},
				},
				{
					InstanceID: 1,
					CumulativeUsage: reporter.CumulativeUsage{
						Value: 0.8,
					},
				},
				{
					InstanceID: 2,
					CumulativeUsage: reporter.CumulativeUsage{
						Value: 0.875,
					},
				},
			},
		}

		instanceReporter.CreateApplicationReportReturns(applicationReport, nil)
	})

	JustBeforeEach(func() {
		runResult = runner.Run("app-name")
	})

	It("prints the app CPU metrics", func() {
		Expect(runResult.IsFailure).To(BeFalse())

		Expect(instanceReporter.CreateApplicationReportCallCount()).To(Equal(1))
		actualAppName := instanceReporter.CreateApplicationReportArgsForCall(0)
		Expect(actualAppName).To(Equal("app-name"))

		Expect(outputRenderer.ShowApplicationReportCallCount()).To(Equal(1))
		actualApplicationReport := outputRenderer.ShowApplicationReportArgsForCall(0)
		Expect(actualApplicationReport).To(Equal(applicationReport))
	})

	When("creating the reports fails with a unsupported cf-deployment error", func() {
		BeforeEach(func() {
			instanceReporter.CreateApplicationReportReturns(reporter.ApplicationReport{}, reporter.NewUnsupportedCFDeploymentError("app-name"))
		})

		It("returns a failure", func() {
			Expect(runResult.IsFailure).To(BeTrue())
			Expect(runResult.ErrorMessage).To(ContainSubstring("app-name"))
			Expect(runResult.WarningMessage).To(BeEmpty())
		})
	})

	When("creating the reports fails with a general error", func() {
		BeforeEach(func() {
			instanceReporter.CreateApplicationReportReturns(reporter.ApplicationReport{}, errors.New("reports error"))
		})

		It("returns a failure", func() {
			Expect(runResult.IsFailure).To(BeTrue())
			Expect(runResult.ErrorMessage).To(Equal("reports error"))
			Expect(runResult.WarningMessage).To(ContainSubstring("Your Cloud Foundry may not have enabled the CPU Entitlements feature. Please consult your operator."))
		})
	})

	When("rendering the app metrics fails", func() {
		BeforeEach(func() {
			outputRenderer.ShowApplicationReportReturns(errors.New("render error"))
		})

		It("returns a failure", func() {
			Expect(runResult.IsFailure).To(BeTrue())
			Expect(runResult.ErrorMessage).To(Equal("render error"))
		})
	})
})
