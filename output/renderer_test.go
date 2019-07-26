package output_test

import (
	"github.com/fatih/color"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cli/cf/terminal"
	models "code.cloudfoundry.org/cli/plugin/models"
	"code.cloudfoundry.org/cpu-entitlement-plugin/calculator"
	"code.cloudfoundry.org/cpu-entitlement-plugin/metadata"
	"code.cloudfoundry.org/cpu-entitlement-plugin/output"
	"code.cloudfoundry.org/cpu-entitlement-plugin/output/outputfakes"
)

var _ = Describe("Renderer", func() {
	var (
		appInfo         metadata.CFAppInfo
		instanceReports []calculator.InstanceReport
		display         *outputfakes.FakeDisplay
		renderer        output.Renderer
	)

	BeforeEach(func() {
		appInfo = metadata.CFAppInfo{
			App:      models.GetAppModel{Name: "myapp"},
			Username: "theuser",
			Org:      "theorg",
			Space:    "thespace",
		}
		instanceReports = []calculator.InstanceReport{
			{
				InstanceId:       123,
				EntitlementUsage: 0.5,
			},
			{
				InstanceId:       432,
				EntitlementUsage: 0.75,
			},
		}

		display = new(outputfakes.FakeDisplay)
		renderer = output.NewRenderer(display)
	})

	Describe("ShowMetrics", func() {
		JustBeforeEach(func() {
			renderer.ShowInstanceReports(appInfo, instanceReports)
		})

		It("shows a message with the application info", func() {
			Expect(display.ShowMessageCallCount()).To(Equal(1))
			message, values := display.ShowMessageArgsForCall(0)
			Expect(message).To(Equal("Showing CPU usage against entitlement for app %s in org %s / space %s as %s ...\n"))
			Expect(values).To(Equal([]interface{}{
				terminal.EntityNameColor("myapp"),
				terminal.EntityNameColor("theorg"),
				terminal.EntityNameColor("thespace"),
				terminal.EntityNameColor("theuser"),
			}))
		})

		It("shows the instances table", func() {
			Expect(display.ShowTableCallCount()).To(Equal(1))
			headers, rows := display.ShowTableArgsForCall(0)
			Expect(headers).To(Equal([]string{"", terminal.Colorize("usage", color.Bold)}))
			Expect(rows).To(Equal([][]string{
				{"#123", "50.00%"},
				{"#432", "75.00%"},
			}))
		})

		When("one or more of the instances is above entitlement", func() {
			BeforeEach(func() {
				instanceReports[1].EntitlementUsage = 1.5
			})

			It("highlights the overentitled row", func() {
				Expect(display.ShowTableCallCount()).To(Equal(1))
				_, rows := display.ShowTableArgsForCall(0)
				Expect(rows).To(Equal([][]string{
					{"#123", "50.00%"},
					{terminal.Colorize("#432", color.FgRed), terminal.Colorize("150.00%", color.FgRed)},
				}))
			})

			It("prints a tip about overentitlement", func() {
				Expect(display.ShowMessageCallCount()).To(Equal(2))
				message, _ := display.ShowMessageArgsForCall(1)
				Expect(message).To(Equal(terminal.Colorize("\nTIP: Some instances are over their CPU entitlement. Consider scaling your memory or instances.", color.FgCyan)))
			})
		})

		When("one of the instances is between 95% and 100% entitlement", func() {
			BeforeEach(func() {
				instanceReports[1].EntitlementUsage = 0.96
			})

			It("highlights the near overentitled row", func() {
				Expect(display.ShowTableCallCount()).To(Equal(1))
				_, rows := display.ShowTableArgsForCall(0)
				Expect(rows).To(Equal([][]string{
					{"#123", "50.00%"},
					{terminal.Colorize("#432", color.FgYellow), terminal.Colorize("96.00%", color.FgYellow)},
				}))
			})

			It("prints a tip about near overentitlement", func() {
				Expect(display.ShowMessageCallCount()).To(Equal(2))
				message, _ := display.ShowMessageArgsForCall(1)
				Expect(message).To(Equal(terminal.Colorize("\nTIP: Some instances are near their CPU entitlement. Consider scaling your memory or instances.", color.FgCyan)))
			})
		})

		When("one of the instances is between 95% and 100% entitlement, and one is over 100%", func() {
			BeforeEach(func() {
				instanceReports[0].EntitlementUsage = 0.96
				instanceReports[1].EntitlementUsage = 1.5
				instanceReports = append(instanceReports, instanceReports[0])
			})

			It("highlights both rows in various colours", func() {
				Expect(display.ShowTableCallCount()).To(Equal(1))
				_, rows := display.ShowTableArgsForCall(0)
				Expect(rows).To(Equal([][]string{
					{terminal.Colorize("#123", color.FgYellow), terminal.Colorize("96.00%", color.FgYellow)},
					{terminal.Colorize("#432", color.FgRed), terminal.Colorize("150.00%", color.FgRed)},
					{terminal.Colorize("#123", color.FgYellow), terminal.Colorize("96.00%", color.FgYellow)},
				}))
			})

			It("prints a tip about over overentitlement and not about near overentitlement", func() {
				Expect(display.ShowMessageCallCount()).To(Equal(2))
				message, _ := display.ShowMessageArgsForCall(1)
				Expect(message).To(Equal(terminal.Colorize("\nTIP: Some instances are over their CPU entitlement. Consider scaling your memory or instances.", color.FgCyan)))
			})
		})
	})
})
