package output

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cpu-entitlement-plugin/reporter"
)

type OverEntitlementInstancesRenderer struct{}

func NewOverEntitlementInstancesRenderer(ui terminal.UI) *OverEntitlementInstancesRenderer {
	return &OverEntitlementInstancesRenderer{}
}

func (r *OverEntitlementInstancesRenderer) Render(report reporter.OEIReport) error {
	if len(report.SpaceReports) == 0 {
		fmt.Printf("No apps over entitlement")
	} else {
		fmt.Printf("Report = %+v\n", report)
	}
	return nil
}
