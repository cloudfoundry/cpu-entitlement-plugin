package main

import (
	"code.cloudfoundry.org/cli/plugin"
	"code.cloudfoundry.org/cpu-entitlement-plugin/plugins/cpu_overentitlement_instances"
)

func main() {
	plugin.Start(cpu_overentitlement_instances.New())
}
