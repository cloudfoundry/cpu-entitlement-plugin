package output

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cpu-entitlement-admin-plugin/reporter"
)

type Renderer struct{}

func NewRenderer(ui terminal.UI) *Renderer {
	return &Renderer{}
}

func (r *Renderer) Render(report reporter.Report) error {
	if len(report.SpaceReports) == 0 {
		fmt.Printf("No apps over entitlement")
	} else {
		fmt.Printf("Report = %+v\n", report)
	}
	return nil
}
