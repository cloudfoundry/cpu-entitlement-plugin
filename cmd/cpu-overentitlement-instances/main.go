package main

import (
	"code.cloudfoundry.org/cli/plugin"
	"code.cloudfoundry.org/cpu-entitlement-plugin/plugins"
)

func main() {
	plugin.Start(plugins.NewOverEntitlementInstancesPlugin())
}
