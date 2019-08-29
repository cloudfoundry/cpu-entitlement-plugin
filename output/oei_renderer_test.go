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
		report = reporter.OEIReport{Org: "org", Username: "user"}
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

})
