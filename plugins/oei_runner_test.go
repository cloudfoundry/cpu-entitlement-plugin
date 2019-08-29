package plugins_test

import (
	"errors"

	"code.cloudfoundry.org/cpu-entitlement-plugin/plugins"
	"code.cloudfoundry.org/cpu-entitlement-plugin/plugins/pluginsfakes"
	"code.cloudfoundry.org/cpu-entitlement-plugin/reporter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Runner", func() {
	var (
		fakeReporter *pluginsfakes.FakeOverEntitlementInstancesReporter
		fakeRenderer *pluginsfakes.FakeOverEntitlementInstancesRenderer

		runner *plugins.OverEntitlementInstancesRunner
		err    error
		report reporter.OEIReport
	)

	BeforeEach(func() {
		fakeReporter = new(pluginsfakes.FakeOverEntitlementInstancesReporter)
		fakeRenderer = new(pluginsfakes.FakeOverEntitlementInstancesRenderer)

		report = reporter.OEIReport{
			SpaceReports: []reporter.SpaceReport{
				{
					SpaceName: "space-1",
					Apps: []string{
						"app-1",
						"app-2",
					},
				}, {
					SpaceName: "space-2",
					Apps: []string{
						"app-1",
					},
				},
			},
		}

		fakeReporter.OverEntitlementInstancesReturns(report, nil)

		runner = plugins.NewOverEntitlementInstancesRunner(fakeReporter, fakeRenderer)
	})

	JustBeforeEach(func() {
		err = runner.Run()
	})

	It("collects reports and renders them", func() {
		Expect(err).NotTo(HaveOccurred())
		Expect(fakeRenderer.RenderCallCount()).To(Equal(1))
		Expect(fakeRenderer.RenderArgsForCall(0)).To(Equal(report))
	})

	When("the reporter fails", func() {
		BeforeEach(func() {
			fakeReporter.OverEntitlementInstancesReturns(reporter.OEIReport{}, errors.New("reporter-err"))
		})

		It("returns the error", func() {
			Expect(err).To(MatchError("reporter-err"))
		})
	})

	When("the renderer fails", func() {
		BeforeEach(func() {
			fakeRenderer.RenderReturns(errors.New("renderer-err"))
		})

		It("returns the error", func() {
			Expect(err).To(MatchError("renderer-err"))
		})
	})
})
