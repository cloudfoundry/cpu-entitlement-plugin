package output

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cpu-entitlement-plugin/cf"
	"code.cloudfoundry.org/cpu-entitlement-plugin/reporter"
	"github.com/fatih/color"
)

const DateFmt = "2006-01-02 15:04:05"
const noColor color.Attribute = -1

type AppRenderer struct {
	display Display
}

//go:generate counterfeiter . Display

type Display interface {
	ShowMessage(message string, values ...interface{})
	ShowTable(headers []string, rows [][]string) error
}

func NewAppRenderer(display Display) AppRenderer {
	return AppRenderer{display: display}
}

func (r AppRenderer) ShowInstanceReports(application cf.Application, instanceReports []reporter.InstanceReport) error {
	r.showAppInfoHeader(application)

	var rows [][]string

	var status string
	var reportsWithSpikes []reporter.InstanceReport
	for _, report := range instanceReports {
		rowColor := noColor
		instanceID := fmt.Sprintf("#%d", report.InstanceID)
		avgEntitlementRatio := fmt.Sprintf("%.2f%%", report.HistoricalUsage.Value*100)
		if report.HistoricalUsage.Value > 1 {
			status = "over"
			rowColor = color.FgRed
		} else if report.HistoricalUsage.Value > 0.95 {
			if status == "" {
				status = "near"
			}
			rowColor = color.FgYellow
		}

		if report.HasRecordedSpike() && report.HistoricalUsage.Value <= 1 {
			reportsWithSpikes = append(reportsWithSpikes, report)
		}

		currEntitlementRatio := fmt.Sprintf("%.2f%%", report.CurrentUsage.Value*100)

		rows = append(rows, colorizeRow([]string{instanceID, avgEntitlementRatio, currEntitlementRatio}, rowColor))
	}

	err := r.display.ShowTable([]string{"", terminal.Colorize("avg usage", color.Bold), terminal.Colorize("curr usage", color.Bold)}, rows)
	if err != nil {
		return err
	}

	if status != "" {
		r.display.ShowMessage(terminal.Colorize(fmt.Sprintf("TIP: Some instances are %s their CPU entitlement. Consider scaling your memory or instances.", status), color.FgCyan))
	}

	for _, reportWithSpike := range reportsWithSpikes {
		historicalUsage := reportWithSpike.HistoricalUsage
		if historicalUsage.LastSpikeFrom.Equal(historicalUsage.LastSpikeTo) {
			r.display.ShowMessage(terminal.Colorize(fmt.Sprintf("WARNING: Instance #%d was over entitlement at %s", reportWithSpike.InstanceID, historicalUsage.LastSpikeFrom.Format(DateFmt)), color.FgYellow))
		} else {
			r.display.ShowMessage(terminal.Colorize(fmt.Sprintf("WARNING: Instance #%d was over entitlement from %s to %s", reportWithSpike.InstanceID, historicalUsage.LastSpikeFrom.Format(DateFmt), historicalUsage.LastSpikeTo.Format(DateFmt)), color.FgYellow))
		}
	}

	return nil
}

func (r AppRenderer) ShowMessage(application cf.Application, message string, values ...interface{}) {
	r.showAppInfoHeader(application)
	r.display.ShowMessage(message, values...)
}

func (r AppRenderer) showAppInfoHeader(application cf.Application) {
	r.display.ShowMessage("Showing CPU usage against entitlement for app %s in org %s / space %s as %s ...\n",
		terminal.EntityNameColor(application.Name),
		terminal.EntityNameColor(application.Org),
		terminal.EntityNameColor(application.Space),
		terminal.EntityNameColor(application.Username),
	)
}

func colorizeRow(row []string, rowColor color.Attribute) []string {
	if rowColor == noColor {
		return row
	}

	colorizedRow := []string{}
	for _, col := range row {
		colorizedRow = append(colorizedRow, terminal.Colorize(col, rowColor))
	}

	return colorizedRow
}
