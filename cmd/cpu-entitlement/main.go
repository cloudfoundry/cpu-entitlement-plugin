package main

import (
	"code.cloudfoundry.org/cli/plugin"
	"code.cloudfoundry.org/cpu-entitlement-plugin/plugins/cpu_entitlement"
)

func main() {
	plugin.Start(cpu_entitlement.New())
}
