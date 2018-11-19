package main

import (
	"code.cloudfoundry.org/cli/plugin"
	cpuplugin "github.com/cloudfoundry/cpu-entitlement-plugin/plugin"
)

func main() {
	plugin.Start(cpuplugin.New())
}
