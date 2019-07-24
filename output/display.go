package output

import "code.cloudfoundry.org/cli/cf/terminal"

type TerminalDisplay struct {
	ui terminal.UI
}

func NewTerminalDisplay(ui terminal.UI) TerminalDisplay {
	return TerminalDisplay{ui: ui}
}

func (d TerminalDisplay) ShowMessage(message string, values ...interface{}) {
	d.ui.Say(message, values...)
}

func (d TerminalDisplay) ShowTable(headers []string, rows [][]string) {
	table := d.ui.Table(headers)
	for _, row := range rows {
		table.Add(row...)
	}
	table.Print()
}
