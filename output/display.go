package output

import (
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/lager"
)

type TerminalDisplay struct {
	ui terminal.UI
}

func NewTerminalDisplay(ui terminal.UI) TerminalDisplay {
	return TerminalDisplay{ui: ui}
}

func (d TerminalDisplay) ShowMessage(message string, values ...interface{}) {
	d.ui.Say(message, values...)
}

func (d TerminalDisplay) ShowTable(logger lager.Logger, headers []string, rows [][]string) error {
	logger = logger.Session("terminal-display-show-table")

	table := d.ui.Table(headers)
	for _, row := range rows {
		table.Add(row...)
	}

	err := table.Print()
	if err != nil {
		logger.Error("table-print-failed", err)
		return err
	}

	d.ui.Say("")

	return nil
}
