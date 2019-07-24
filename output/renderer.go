package output

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cpu-entitlement-plugin/metadata"
	"code.cloudfoundry.org/cpu-entitlement-plugin/usagemetric"
	"github.com/fatih/color"
)

type Renderer struct {
	display Display
}

//go:generate counterfeiter . Display

type Display interface {
	ShowMessage(message string, values ...interface{})
	ShowTable(headers []string, rows [][]string)
}

func NewRenderer(display Display) Renderer {
	return Renderer{display: display}
}

func (r Renderer) ShowMetrics(info metadata.CFAppInfo, metrics []usagemetric.UsageMetric) {
	r.display.ShowMessage("Showing CPU usage against entitlement for app %s in org %s / space %s as %s ...\n",
		terminal.EntityNameColor(info.App.Name),
		terminal.EntityNameColor(info.Org),
		terminal.EntityNameColor(info.Space),
		terminal.EntityNameColor(info.Username),
	)

	var rows [][]string

	overentitled := false
	for _, metric := range metrics {
		instanceId := fmt.Sprintf("#%d", metric.InstanceId)
		cpuUsage := fmt.Sprintf("%.2f%%", metric.CPUUsage()*100)
		if metric.CPUUsage() > 1 {
			overentitled = true
			instanceId = terminal.Colorize(instanceId, color.FgRed)
			cpuUsage = terminal.Colorize(cpuUsage, color.FgRed)
		}

		rows = append(rows, []string{instanceId, cpuUsage})
	}

	r.display.ShowTable([]string{"", bold("usage")}, rows)

	if overentitled {
		r.display.ShowMessage(terminal.Colorize("TIP: Some instances are over their CPU entitlement. Consider scaling your memory or instances.", color.FgCyan))
	}
}

func bold(message string) string {
	return terminal.Colorize(message, color.Bold)
}
