package main

import (
	"code.cloudfoundry.org/cli/plugin"
	"code.cloudfoundry.org/cpu-entitlement-plugin/plugins/app"
)

func main() {
	plugin.Start(app.New())
}
