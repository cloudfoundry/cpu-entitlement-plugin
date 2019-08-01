package output

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cpu-entitlement-plugin/calculator"
	"code.cloudfoundry.org/cpu-entitlement-plugin/metadata"
	"github.com/fatih/color"
)

const DateFmt = "2006-01-02 15:04:05"

type Renderer struct {
	display Display
}

//go:generate counterfeiter . Display

type Display interface {
	ShowMessage(message string, values ...interface{})
	ShowTable(headers []string, rows [][]string) error
}

func NewRenderer(display Display) Renderer {
	return Renderer{display: display}
}

func (r Renderer) ShowInstanceReports(info metadata.CFAppInfo, instanceReports []calculator.InstanceReport) error {
	r.display.ShowMessage("Showing CPU usage against entitlement for app %s in org %s / space %s as %s ...\n",
		terminal.EntityNameColor(info.App.Name),
		terminal.EntityNameColor(info.Org),
		terminal.EntityNameColor(info.Space),
		terminal.EntityNameColor(info.Username),
	)

	var rows [][]string

	var status string
	var reportsWithSpikes []calculator.InstanceReport
	for _, report := range instanceReports {
		instanceID := fmt.Sprintf("#%d", report.InstanceID)
		entitlementRatio := fmt.Sprintf("%.2f%%", report.EntitlementUsage*100)
		if report.EntitlementUsage > 1 {
			status = "over"
			instanceID = terminal.Colorize(instanceID, color.FgRed)
			entitlementRatio = terminal.Colorize(entitlementRatio, color.FgRed)
		} else if report.EntitlementUsage > 0.95 {
			if status == "" {
				status = "near"
			}
			instanceID = terminal.Colorize(instanceID, color.FgYellow)
			entitlementRatio = terminal.Colorize(entitlementRatio, color.FgYellow)
		}

		if report.HasRecordedSpike() && report.EntitlementUsage <= 1 {
			reportsWithSpikes = append(reportsWithSpikes, report)
		}

		rows = append(rows, []string{instanceID, entitlementRatio})
	}

	err := r.display.ShowTable([]string{"", terminal.Colorize("usage", color.Bold)}, rows)
	if err != nil {
		return err
	}

	if status != "" {
		r.display.ShowMessage(terminal.Colorize(fmt.Sprintf("TIP: Some instances are %s their CPU entitlement. Consider scaling your memory or instances.", status), color.FgCyan))
	}

	for _, reportWithSpike := range reportsWithSpikes {
		if reportWithSpike.LastSpikeFrom.Equal(reportWithSpike.LastSpikeTo) {
			r.display.ShowMessage(terminal.Colorize(fmt.Sprintf("WARNING: Instance #%d was over entitlement at %s", reportWithSpike.InstanceID, reportWithSpike.LastSpikeFrom.Format(DateFmt)), color.FgYellow))
		} else {
			r.display.ShowMessage(terminal.Colorize(fmt.Sprintf("WARNING: Instance #%d was over entitlement from %s to %s", reportWithSpike.InstanceID, reportWithSpike.LastSpikeFrom.Format(DateFmt), reportWithSpike.LastSpikeTo.Format(DateFmt)), color.FgYellow))
		}
	}

	return nil
}
