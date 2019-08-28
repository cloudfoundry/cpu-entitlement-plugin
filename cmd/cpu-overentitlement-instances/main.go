package main

import (
	"code.cloudfoundry.org/cli/plugin"
	"code.cloudfoundry.org/cpu-entitlement-plugin/plugins/org"
)

func main() {
	plugin.Start(org.New())
}
