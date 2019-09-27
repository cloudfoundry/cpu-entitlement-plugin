package output_test

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cpu-entitlement-plugin/output"
	"code.cloudfoundry.org/cpu-entitlement-plugin/output/outputfakes"
	"code.cloudfoundry.org/cpu-entitlement-plugin/reporter"
)

var _ = Describe("Renderer", func() {
	var (
		instanceReports []reporter.InstanceReport
		display         *outputfakes.FakeAppDisplay
		renderer        output.AppRenderer
	)

	BeforeEach(func() {
		instanceReports = []reporter.InstanceReport{
			{
				InstanceID: 123,
				HistoricalUsage: reporter.HistoricalUsage{
					Value: 0.5,
				},
				CurrentUsage: reporter.CurrentUsage{
					Value: 1.5,
				},
			},
			{
				InstanceID: 432,
				HistoricalUsage: reporter.HistoricalUsage{
					Value: 0.75,
				},
				CurrentUsage: reporter.CurrentUsage{
					Value: 1.75,
				},
			},
		}

		display = new(outputfakes.FakeAppDisplay)
		renderer = output.NewAppRenderer(display)
	})

	Describe("ShowMetrics", func() {
		var (
			appReport reporter.ApplicationReport
		)
		JustBeforeEach(func() {
			appReport = reporter.ApplicationReport{ApplicationName: "myapp", Org: "theorg", Space: "thespace", Username: "theuser", InstanceReports: instanceReports}
			Expect(renderer.ShowApplicationReport(appReport)).To(Succeed())
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
			Expect(headers).To(Equal([]string{"", bold("avg usage"), bold("curr usage")}))
			Expect(rows).To(Equal([][]string{
				{"#123", "50.00%", "150.00%"},
				{"#432", "75.00%", "175.00%"},
			}))
		})

		When("there are no instances of the application", func() {
			BeforeEach(func() {
				instanceReports = []reporter.InstanceReport{}
			})

			It("prints a message about no instances", func() {
				Expect(display.ShowTableCallCount()).To(Equal(0))
				Expect(display.ShowMessageCallCount()).To(Equal(2))
				message, values := display.ShowMessageArgsForCall(1)
				Expect(message).To(Equal("There are no running instances of this application."))
				Expect(values).To(BeEmpty())
			})
		})

		When("one or more of the instances is above entitlement", func() {
			BeforeEach(func() {
				instanceReports[1].HistoricalUsage.Value = 1.5
			})

			It("highlights the overentitled row", func() {
				Expect(display.ShowTableCallCount()).To(Equal(1))
				_, rows := display.ShowTableArgsForCall(0)
				Expect(rows).To(Equal([][]string{
					{"#123", "50.00%", "150.00%"},
					redRow("#432", "150.00%", "175.00%"),
				}))
			})

			It("prints a tip about overentitlement", func() {
				Expect(display.ShowMessageCallCount()).To(Equal(2))
				message, _ := display.ShowMessageArgsForCall(1)
				Expect(message).To(Equal(cyan("WARNING: Some instances are over their CPU entitlement. Consider scaling your memory or instances.")))
			})
		})

		When("one of the instances is between 95% and 100% entitlement", func() {
			BeforeEach(func() {
				instanceReports[1].HistoricalUsage.Value = 0.96
			})

			It("highlights the near overentitled row", func() {
				Expect(display.ShowTableCallCount()).To(Equal(1))
				_, rows := display.ShowTableArgsForCall(0)
				Expect(rows).To(Equal([][]string{
					{"#123", "50.00%", "150.00%"},
					yellowRow("#432", "96.00%", "175.00%"),
				}))
			})

			It("prints a tip about near overentitlement", func() {
				Expect(display.ShowMessageCallCount()).To(Equal(2))
				message, _ := display.ShowMessageArgsForCall(1)
				Expect(message).To(Equal(cyan("TIP: Some instances are near their CPU entitlement. Consider scaling your memory or instances.")))
			})
		})

		When("one of the instances is between 95% and 100% entitlement, and one is over 100%", func() {
			BeforeEach(func() {
				instanceReports[0].HistoricalUsage.Value = 0.96
				instanceReports[1].HistoricalUsage.Value = 1.5
				instanceReports = append(instanceReports, instanceReports[0])
			})

			It("highlights both rows in various colours", func() {
				Expect(display.ShowTableCallCount()).To(Equal(1))
				_, rows := display.ShowTableArgsForCall(0)
				Expect(rows).To(Equal([][]string{
					yellowRow("#123", "96.00%", "150.00%"),
					redRow("#432", "150.00%", "175.00%"),
					yellowRow("#123", "96.00%", "150.00%"),
				}))
			})

			It("prints a tip about over overentitlement and not about near overentitlement", func() {
				Expect(display.ShowMessageCallCount()).To(Equal(2))
				message, _ := display.ShowMessageArgsForCall(1)
				Expect(message).To(Equal(cyan("WARNING: Some instances are over their CPU entitlement. Consider scaling your memory or instances.")))
			})
		})

		When("one or more instances have been over entitlement", func() {
			BeforeEach(func() {
				instanceReports = append(instanceReports, reporter.InstanceReport{
					InstanceID: 234,
					HistoricalUsage: reporter.HistoricalUsage{
						Value:         0.5,
						LastSpikeFrom: time.Date(2019, 7, 30, 9, 0, 0, 0, time.UTC),
						LastSpikeTo:   time.Date(2019, 7, 31, 12, 0, 0, 0, time.UTC),
					},
				},
					reporter.InstanceReport{
						InstanceID: 345,
						HistoricalUsage: reporter.HistoricalUsage{
							Value:         0.5,
							LastSpikeFrom: time.Date(2019, 6, 15, 10, 0, 0, 0, time.UTC),
							LastSpikeTo:   time.Date(2019, 6, 21, 5, 0, 0, 0, time.UTC),
						},
					})
			})

			It("prints warnings about instances having been over entitlement", func() {
				Expect(display.ShowMessageCallCount()).To(Equal(3))
				firstWarning, _ := display.ShowMessageArgsForCall(1)
				Expect(firstWarning).To(Equal(yellow(fmt.Sprintf("WARNING: Instance #234 was over entitlement from 2019-07-30 09:00:00 to 2019-07-31 12:00:00"))))
				secondWarning, _ := display.ShowMessageArgsForCall(2)
				Expect(secondWarning).To(Equal(yellow(fmt.Sprintf("WARNING: Instance #345 was over entitlement from 2019-06-15 10:00:00 to 2019-06-21 05:00:00"))))
			})
		})

		When("an instance is currently over entitlement with a 'current' spike", func() {
			BeforeEach(func() {
				instanceReports = append(instanceReports, reporter.InstanceReport{
					InstanceID: 234,
					HistoricalUsage: reporter.HistoricalUsage{
						Value:         1.5,
						LastSpikeFrom: time.Date(2019, 7, 30, 9, 0, 0, 0, time.UTC),
						LastSpikeTo:   time.Date(2019, 7, 31, 12, 0, 0, 0, time.UTC),
					},
				})
			})

			It("suppresses warning about instance having been over entitlement", func() {
				Expect(display.ShowMessageCallCount()).To(Equal(2))
			})
		})

		When("spike was instantaneous", func() {
			BeforeEach(func() {
				instanceReports = append(instanceReports, reporter.InstanceReport{
					InstanceID: 234,
					HistoricalUsage: reporter.HistoricalUsage{
						Value:         0.5,
						LastSpikeFrom: time.Date(2019, 7, 31, 12, 0, 0, 0, time.UTC),
						LastSpikeTo:   time.Date(2019, 7, 31, 12, 0, 0, 0, time.UTC),
					},
				})
			})

			It("says 'at', not 'from'...'to' in the warning message", func() {
				Expect(display.ShowMessageCallCount()).To(Equal(2))
				warning, _ := display.ShowMessageArgsForCall(1)
				Expect(warning).To(Equal(yellow(fmt.Sprintf("WARNING: Instance #234 was over entitlement at 2019-07-31 12:00:00"))))
			})
		})
	})
})

func yellow(s string) string {
	return terminal.Colorize(s, color.FgYellow)
}

func cyan(s string) string {
	return terminal.Colorize(s, color.FgCyan)
}

func bold(s string) string {
	return terminal.Colorize(s, color.Bold)
}

func colorizeRow(row []string, rowColor color.Attribute) []string {
	colorizedRow := []string{}
	for _, col := range row {
		colorizedRow = append(colorizedRow, terminal.Colorize(col, rowColor))
	}

	return colorizedRow
}

func yellowRow(r ...string) []string {
	return colorizeRow(r, color.FgYellow)
}

func redRow(r ...string) []string {
	return colorizeRow(r, color.FgRed)
}
