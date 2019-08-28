package org_test

import (
	"errors"

	plugin "code.cloudfoundry.org/cpu-entitlement-plugin/plugins/org"
	"code.cloudfoundry.org/cpu-entitlement-plugin/plugins/org/orgfakes"
	"code.cloudfoundry.org/cpu-entitlement-plugin/reporter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Runner", func() {
	var (
		fakeReporter *orgfakes.FakeReporter
		fakeRenderer *orgfakes.FakeRenderer

		runner *plugin.Runner
		err    error
		report reporter.OEIReport
	)

	BeforeEach(func() {
		fakeReporter = new(orgfakes.FakeReporter)
		fakeRenderer = new(orgfakes.FakeRenderer)

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

		runner = plugin.NewRunner(fakeReporter, fakeRenderer)
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
