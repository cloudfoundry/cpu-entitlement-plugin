package output_test

import (
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cpu-entitlement-plugin/output"
	"code.cloudfoundry.org/cpu-entitlement-plugin/output/outputfakes"
	"code.cloudfoundry.org/cpu-entitlement-plugin/reporter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OEI Renderer", func() {
	var (
		display   *outputfakes.FakeOverEntitlementInstancesDisplay
		report    reporter.OEIReport
		renderErr error
		renderer  *output.OverEntitlementInstancesRenderer
	)

	BeforeEach(func() {
		display = new(outputfakes.FakeOverEntitlementInstancesDisplay)
		spaceReports := []reporter.SpaceReport{
			reporter.SpaceReport{SpaceName: "space-1", Apps: []string{"app-1-1", "app-1-2"}},
			reporter.SpaceReport{SpaceName: "space-2", Apps: []string{"app-2-1"}},
		}
		report = reporter.OEIReport{Org: "org", Username: "user", SpaceReports: spaceReports}
		renderer = output.NewOverEntitlementInstancesRenderer(display)
	})

	JustBeforeEach(func() {
		renderErr = renderer.Render(report)
	})

	It("succeeds", func() {
		Expect(renderErr).NotTo(HaveOccurred())
	})

	It("shows report header", func() {
		Expect(display.ShowMessageCallCount()).To(Equal(1))
		actualMsg, actualMsgArgs := display.ShowMessageArgsForCall(0)
		Expect(actualMsg).To(Equal("Showing over-entitlement apps in org %s as %s...\n"))
		Expect(actualMsgArgs).To(ConsistOf(terminal.EntityNameColor("org"), terminal.EntityNameColor("user")))
	})

	It("shows applications over entitlement", func() {
		Expect(display.ShowTableCallCount()).To(Equal(1))
		headers, rows := display.ShowTableArgsForCall(0)
		Expect(headers).To(ConsistOf("space", "app"))

		Expect(len(rows)).To(Equal(3))
		Expect(rows[0]).To(ConsistOf("space-1", "app-1-1"))
		Expect(rows[1]).To(ConsistOf("space-1", "app-1-2"))
		Expect(rows[2]).To(ConsistOf("space-2", "app-2-1"))
	})

	When("there are no applications over entitlement", func() {
		BeforeEach(func() {
			report = reporter.OEIReport{Org: "org", Username: "user"}
		})

		It("shows a no applications over entitlement message", func() {
			Expect(display.ShowMessageCallCount()).To(Equal(1))
			actualMsg, actualMsgArgs := display.ShowMessageArgsForCall(0)
			Expect(actualMsg).To(ContainSubstring("No apps over entitlement in org"))
			Expect(actualMsgArgs).To(ConsistOf(terminal.EntityNameColor("org")))
		})
	})

})
