package cpu_entitlement_test

import (
	"errors"

	"code.cloudfoundry.org/cpu-entitlement-plugin/metadata"
	"code.cloudfoundry.org/cpu-entitlement-plugin/plugins/cpu_entitlement"
	"code.cloudfoundry.org/cpu-entitlement-plugin/plugins/cpu_entitlement/cpu_entitlementfakes"
	"code.cloudfoundry.org/cpu-entitlement-plugin/reporter"
	"code.cloudfoundry.org/cpu-entitlement-plugin/result"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Runner", func() {
	var (
		infoGetter       *cpu_entitlementfakes.FakeCFAppInfoGetter
		instanceReporter *cpu_entitlementfakes.FakeReporter
		outputRenderer   *cpu_entitlementfakes.FakeOutputRenderer

		runner    cpu_entitlement.Runner
		runResult result.Result
	)

	BeforeEach(func() {
		infoGetter = new(cpu_entitlementfakes.FakeCFAppInfoGetter)
		instanceReporter = new(cpu_entitlementfakes.FakeReporter)
		outputRenderer = new(cpu_entitlementfakes.FakeOutputRenderer)
		runner = cpu_entitlement.NewRunner(infoGetter, instanceReporter, outputRenderer)

		infoGetter.GetCFAppInfoReturns(metadata.CFAppInfo{
			Guid:      "123",
			Name:      "app-name",
			Instances: map[int]metadata.CFAppInstance{0: metadata.CFAppInstance{}},
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
		actualAppInfo := instanceReporter.CreateInstanceReportsArgsForCall(0)
		Expect(actualAppInfo.Guid).To(Equal("123"))

		Expect(outputRenderer.ShowInstanceReportsCallCount()).To(Equal(1))
		info, instanceReports := outputRenderer.ShowInstanceReportsArgsForCall(0)
		Expect(info).To(Equal(metadata.CFAppInfo{
			Guid:      "123",
			Name:      "app-name",
			Instances: map[int]metadata.CFAppInstance{0: metadata.CFAppInstance{}},
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
			instanceReporter.CreateInstanceReportsReturns([]reporter.InstanceReport{}, nil)
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
