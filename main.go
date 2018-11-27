package main

import (
	"code.cloudfoundry.org/cli/plugin"
	cpuplugin "code.cloudfoundry.org/cpu-entitlement-plugin/plugin"
)

func main() {
	plugin.Start(cpuplugin.New())
}
