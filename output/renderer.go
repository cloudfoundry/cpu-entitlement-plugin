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
	for _, metric := range metrics {
		rows = append(rows, []string{fmt.Sprintf("#%d", metric.InstanceId), fmt.Sprintf("%.2f%%", metric.CPUUsage()*100)})
	}

	r.display.ShowTable([]string{"", bold("usage")}, rows)
}

func bold(message string) string {
	return terminal.Colorize(message, color.Bold)
}
