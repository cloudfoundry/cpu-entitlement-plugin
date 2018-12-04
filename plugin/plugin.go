package plugin // import "code.cloudfoundry.org/cpu-entitlement-plugin/plugin"

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"

	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/cf/trace"
	"code.cloudfoundry.org/cli/plugin"
	"code.cloudfoundry.org/cpu-entitlement-plugin/logstreamer"
	"code.cloudfoundry.org/cpu-entitlement-plugin/token"
	"github.com/fatih/color"
)

type CPUEntitlementPlugin struct{}

func New() *CPUEntitlementPlugin {
	return &CPUEntitlementPlugin{}
}

func (p *CPUEntitlementPlugin) Run(cli plugin.CliConnection, args []string) {
	if args[0] == "CLI-MESSAGE-UNINSTALL" {
		os.Exit(0)
	}

	traceLogger := trace.NewLogger(os.Stdout, true, os.Getenv("CF_TRACE"), "")
	ui := terminal.NewUI(os.Stdin, os.Stdout, terminal.NewTeePrinter(os.Stdout), traceLogger)

	if len(args) != 2 {
		ui.Failed("Usage: `cf cpu-entitlement APP_NAME`")
		os.Exit(1)
	}

	appName := args[1]

	app, err := cli.GetApp(appName)
	if err != nil {
		ui.Failed(err.Error())
		os.Exit(1)
	}

	dopplerURL, err := cli.DopplerEndpoint()
	if err != nil {
		ui.Failed(err.Error())
		os.Exit(1)
	}

	logStreamURL, err := buildLogStreamURL(dopplerURL)
	if err != nil {
		ui.Failed(err.Error())
		os.Exit(1)
	}

	tokenGetter := token.NewTokenGetter(cli.AccessToken)
	logStreamer := logstreamer.New(logStreamURL, tokenGetter)

	usageMetricsStream := logStreamer.Stream(app.Guid)
	ui.Say("CPU entitlement usage by instance for %s...\n", terminal.EntityNameColor(appName))
	table := ui.Table([]string{bold("Instance ID"), bold("Usage")})
	table.Print()

	for usageMetric := range usageMetricsStream {
		table.Add(fmt.Sprintf("#%s", usageMetric.InstanceId), fmt.Sprintf("%.2f%%", usageMetric.CPUUsage()*100))
		table.Print()
	}
}

func bold(message string) string {
	return terminal.Colorize(message, color.Bold)
}

func buildLogStreamURL(dopplerURL string) (string, error) {
	logStreamURL, err := url.Parse(dopplerURL)
	if err != nil {
		return "", err
	}

	regex, err := regexp.Compile("doppler(\\S+):443")
	match := regex.FindStringSubmatch(logStreamURL.Host)

	if len(match) != 2 {
		return "", errors.New("Unable to parse log-stream endpoint from doppler URL")
	}

	logStreamURL.Scheme = "http"
	logStreamURL.Host = "log-stream" + match[1]

	return logStreamURL.String(), nil
}
