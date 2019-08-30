package output

import (
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cpu-entitlement-plugin/reporter"
)

//go:generate counterfeiter . OverEntitlementInstancesDisplay
type OverEntitlementInstancesDisplay interface {
	ShowMessage(message string, values ...interface{})
	ShowTable(headers []string, rows [][]string) error
}

type OverEntitlementInstancesRenderer struct {
	display OverEntitlementInstancesDisplay
}

func NewOverEntitlementInstancesRenderer(display OverEntitlementInstancesDisplay) *OverEntitlementInstancesRenderer {
	return &OverEntitlementInstancesRenderer{display: display}
}

func (r *OverEntitlementInstancesRenderer) Render(report reporter.OEIReport) error {
	if len(report.SpaceReports) == 0 {
		r.display.ShowMessage("No apps over entitlement in org %s.\n", terminal.EntityNameColor(report.Org))
		return nil
	}

	r.showReportHeader(report)
	r.display.ShowTable([]string{"space", "app"}, buildOEITableRows(report))
	return nil
}

func (r OverEntitlementInstancesRenderer) showReportHeader(report reporter.OEIReport) {
	r.display.ShowMessage("Showing over-entitlement apps in org %s as %s...\n",
		terminal.EntityNameColor(report.Org),
		terminal.EntityNameColor(report.Username),
	)
}

func buildOEITableRows(report reporter.OEIReport) [][]string {
	var rows [][]string
	for _, spaceReport := range report.SpaceReports {
		for _, app := range spaceReport.Apps {
			rows = append(rows, []string{spaceReport.SpaceName, app})
		}
	}
	return rows
}
