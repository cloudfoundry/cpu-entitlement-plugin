package app_test

import (
	"errors"

	"code.cloudfoundry.org/cpu-entitlement-plugin/metadata"
	plugin "code.cloudfoundry.org/cpu-entitlement-plugin/plugins/app"
	"code.cloudfoundry.org/cpu-entitlement-plugin/plugins/app/appfakes"
	"code.cloudfoundry.org/cpu-entitlement-plugin/reporter/app"
	"code.cloudfoundry.org/cpu-entitlement-plugin/result"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Runner", func() {
	var (
		infoGetter       *appfakes.FakeCFAppInfoGetter
		instanceReporter *appfakes.FakeReporter
		outputRenderer   *appfakes.FakeOutputRenderer

		runner    plugin.Runner
		runResult result.Result
	)

	BeforeEach(func() {
		infoGetter = new(appfakes.FakeCFAppInfoGetter)
		instanceReporter = new(appfakes.FakeReporter)
		outputRenderer = new(appfakes.FakeOutputRenderer)
		runner = plugin.NewRunner(infoGetter, instanceReporter, outputRenderer)

		infoGetter.GetCFAppInfoReturns(metadata.CFAppInfo{
			Guid:      "123",
			Name:      "app-name",
			Instances: map[int]metadata.CFAppInstance{0: metadata.CFAppInstance{}},
		}, nil)

		instanceReporter.CreateInstanceReportsReturns([]app.InstanceReport{
			{
				InstanceID: 0,
				HistoricalUsage: app.HistoricalUsage{
					Value: 0.5,
				},
			},
			{
				InstanceID: 1,
				HistoricalUsage: app.HistoricalUsage{
					Value: 0.8,
				},
			},
			{
				InstanceID: 2,
				HistoricalUsage: app.HistoricalUsage{
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
		actualAppInfo := instanceReporter.CreateInstanceReportsArgsForCall(0)
		Expect(actualAppInfo.Guid).To(Equal("123"))

		Expect(outputRenderer.ShowInstanceReportsCallCount()).To(Equal(1))
		info, instanceReports := outputRenderer.ShowInstanceReportsArgsForCall(0)
		Expect(info).To(Equal(metadata.CFAppInfo{
			Guid:      "123",
			Name:      "app-name",
			Instances: map[int]metadata.CFAppInstance{0: metadata.CFAppInstance{}},
		}))
		Expect(instanceReports).To(Equal([]app.InstanceReport{
			{
				InstanceID: 0,
				HistoricalUsage: app.HistoricalUsage{
					Value: 0.5,
				},
			},
			{
				InstanceID: 1,
				HistoricalUsage: app.HistoricalUsage{
					Value: 0.8,
				},
			},
			{
				InstanceID: 2,
				HistoricalUsage: app.HistoricalUsage{
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

	When("there are zero instances of the application", func() {
		BeforeEach(func() {
			infoGetter.GetCFAppInfoReturns(metadata.CFAppInfo{
				Guid: "123",
				Name: "app-name",
			}, nil)

		})
		It("succeeds", func() {
			Expect(runResult.IsFailure).To(BeFalse())
		})

		It("prints a message", func() {
			Expect(outputRenderer.ShowMessageCallCount()).To(Equal(1))
			info, message, _ := outputRenderer.ShowMessageArgsForCall(0)
			Expect(info).To(Equal(metadata.CFAppInfo{
				Guid: "123",
				Name: "app-name",
			}))
			Expect(message).To(Equal("There are no running instances of this process."))
		})

		It("does not try to generate reports", func() {
			Expect(instanceReporter.CreateInstanceReportsCallCount()).To(Equal(0))
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

	When("no reports are returned", func() {
		BeforeEach(func() {
			instanceReporter.CreateInstanceReportsReturns([]app.InstanceReport{}, nil)
		})

		It("returns a failure", func() {
			Expect(runResult.IsFailure).To(BeTrue())
			Expect(runResult.ErrorMessage).To(ContainSubstring("Could not find any CPU data for app app-name"))
		})
	})

	When("rendering the app metrics fails", func() {
		BeforeEach(func() {
			outputRenderer.ShowInstanceReportsReturns(errors.New("render error"))
		})

		It("returns a failure", func() {
			Expect(runResult.IsFailure).To(BeTrue())
			Expect(runResult.ErrorMessage).To(Equal("render error"))
		})
	})
})
