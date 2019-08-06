package plugin_test

import (
	"errors"
	"time"

	models "code.cloudfoundry.org/cli/plugin/models"
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
			App: models.GetAppModel{
				Guid:          "123",
				Name:          "app-name",
				InstanceCount: 3,
			},
		}, nil)

		instanceReporter.CreateInstanceReportsReturns([]reporter.InstanceReport{
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
		}, nil)
	})

	JustBeforeEach(func() {
		runResult = runner.Run("app-name", time.Unix(123, 0), time.Unix(456, 0))
	})

	It("prints the app CPU metrics", func() {
		Expect(runResult.IsFailure).To(BeFalse())

		Expect(infoGetter.GetCFAppInfoCallCount()).To(Equal(1))
		appName := infoGetter.GetCFAppInfoArgsForCall(0)
		Expect(appName).To(Equal("app-name"))

		Expect(instanceReporter.CreateInstanceReportsCallCount()).To(Equal(1))
		guid, from, to := instanceReporter.CreateInstanceReportsArgsForCall(0)
		Expect(guid).To(Equal("123"))
		Expect(from).To(Equal(time.Unix(123, 0)))
		Expect(to).To(Equal(time.Unix(456, 0)))

		Expect(metricsRenderer.ShowInstanceReportsCallCount()).To(Equal(1))
		info, instanceReports := metricsRenderer.ShowInstanceReportsArgsForCall(0)
		Expect(info).To(Equal(metadata.CFAppInfo{
			App: models.GetAppModel{
				Guid:          "123",
				Name:          "app-name",
				InstanceCount: 3,
			},
		}))
		Expect(instanceReports).To(Equal([]reporter.InstanceReport{
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
