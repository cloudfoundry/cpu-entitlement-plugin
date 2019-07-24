package output

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cpu-entitlement-plugin/metadata"
	"code.cloudfoundry.org/cpu-entitlement-plugin/usagemetric"
	"github.com/fatih/color"
)

type Renderer struct {
	ui terminal.UI
}

func NewRenderer(ui terminal.UI) Renderer {
	return Renderer{ui: ui}
}

func (r Renderer) ShowMetrics(info metadata.CFAppInfo, metrics []usagemetric.UsageMetric) {
	r.ui.Say("Showing CPU usage against entitlement for app %s in org %s / space %s as %s ...\n", terminal.EntityNameColor(info.App.Name), terminal.EntityNameColor(info.Org), terminal.EntityNameColor(info.Space), terminal.EntityNameColor(info.Username))

	table := r.ui.Table([]string{"", bold("usage")})
	for _, usageMetric := range metrics {
		table.Add(fmt.Sprintf("#%d", usageMetric.InstanceId), fmt.Sprintf("%.2f%%", usageMetric.CPUUsage()*100))
	}
	table.Print()
}

func bold(message string) string {
	return terminal.Colorize(message, color.Bold)
}
