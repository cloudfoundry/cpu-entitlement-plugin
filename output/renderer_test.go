package output_test

import (
	"github.com/fatih/color"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cli/cf/terminal"
	models "code.cloudfoundry.org/cli/plugin/models"
	"code.cloudfoundry.org/cpu-entitlement-plugin/metadata"
	"code.cloudfoundry.org/cpu-entitlement-plugin/metrics"
	"code.cloudfoundry.org/cpu-entitlement-plugin/output"
	"code.cloudfoundry.org/cpu-entitlement-plugin/output/outputfakes"
)

var _ = Describe("Renderer", func() {
	var (
		appInfo      metadata.CFAppInfo
		usageMetrics []metrics.Usage
		display      *outputfakes.FakeDisplay
		renderer     output.Renderer
	)

	BeforeEach(func() {
		appInfo = metadata.CFAppInfo{
			App:      models.GetAppModel{Name: "myapp"},
			Username: "theuser",
			Org:      "theorg",
			Space:    "thespace",
		}
		usageMetrics = []metrics.Usage{
			{
				InstanceId:          123,
				AbsoluteUsage:       1.0,
				AbsoluteEntitlement: 2.0,
				ContainerAge:        3.0,
			},
			{
				InstanceId:          432,
				AbsoluteUsage:       1.0,
				AbsoluteEntitlement: 1.0,
				ContainerAge:        3.0,
			},
		}

		display = new(outputfakes.FakeDisplay)
		renderer = output.NewRenderer(display)
	})

	Describe("ShowMetrics", func() {
		JustBeforeEach(func() {
			renderer.ShowMetrics(appInfo, usageMetrics)
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
				{"#432", "100.00%"},
			}))
		})

		When("one of the instances is above entitlement", func() {
			BeforeEach(func() {
				usageMetrics[1].AbsoluteUsage = 1.5
			})

			It("highlights the overentitled row", func() {
				Expect(display.ShowTableCallCount()).To(Equal(1))
				headers, rows := display.ShowTableArgsForCall(0)
				Expect(headers).To(Equal([]string{"", terminal.Colorize("usage", color.Bold)}))
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
	})
})
