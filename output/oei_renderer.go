package output

import (
	"fmt"

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
	r.showReportHeader(report)
	if len(report.SpaceReports) == 0 {
		fmt.Printf("No apps over entitlement")
	} else {
		fmt.Printf("Report = %+v\n", report)
	}
	return nil
}
func (r OverEntitlementInstancesRenderer) showReportHeader(report reporter.OEIReport) {
	r.display.ShowMessage("Showing over-entitlement apps in org %s as %s...\n",
		terminal.EntityNameColor(report.Org),
		terminal.EntityNameColor(report.Username),
	)
}
