package output

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cpu-entitlement-plugin/calculator"
	"code.cloudfoundry.org/cpu-entitlement-plugin/metadata"
	"github.com/fatih/color"
)

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

func (r Renderer) ShowInfos(info metadata.CFAppInfo, infos []calculator.InstanceInfo) error {
	r.display.ShowMessage("Showing CPU usage against entitlement for app %s in org %s / space %s as %s ...\n",
		terminal.EntityNameColor(info.App.Name),
		terminal.EntityNameColor(info.Org),
		terminal.EntityNameColor(info.Space),
		terminal.EntityNameColor(info.Username),
	)

	var rows [][]string

	var status string
	for _, i := range infos {
		instanceId := fmt.Sprintf("#%d", i.InstanceId)
		entitlementRatio := fmt.Sprintf("%.2f%%", i.EntitlementUsage*100)
		if i.EntitlementUsage > 1 {
			status = "over"
			instanceId = terminal.Colorize(instanceId, color.FgRed)
			entitlementRatio = terminal.Colorize(entitlementRatio, color.FgRed)
		} else if i.EntitlementUsage > 0.95 {
			if status == "" {
				status = "near"
			}
			instanceId = terminal.Colorize(instanceId, color.FgYellow)
			entitlementRatio = terminal.Colorize(entitlementRatio, color.FgYellow)
		}

		rows = append(rows, []string{instanceId, entitlementRatio})
	}

	err := r.display.ShowTable([]string{"", bold("usage")}, rows)
	if err != nil {
		return err
	}

	if status != "" {
		r.display.ShowMessage(terminal.Colorize(fmt.Sprintf("\nTIP: Some instances are %s their CPU entitlement. Consider scaling your memory or instances.", status), color.FgCyan))
	}

	return nil
}

func bold(message string) string {
	return terminal.Colorize(message, color.Bold)
}
