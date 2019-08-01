package plugin // import "code.cloudfoundry.org/cpu-entitlement-plugin/plugin"

import (
	"errors"
	"net/url"
	"os"
	"regexp"
	"time"

	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/cf/trace"
	"code.cloudfoundry.org/cli/plugin"
	"code.cloudfoundry.org/cpu-entitlement-plugin/calculator"
	"code.cloudfoundry.org/cpu-entitlement-plugin/metadata"
	"code.cloudfoundry.org/cpu-entitlement-plugin/metrics"
	"code.cloudfoundry.org/cpu-entitlement-plugin/output"
	"code.cloudfoundry.org/cpu-entitlement-plugin/token"
)

type CPUEntitlementPlugin struct{}

func New() CPUEntitlementPlugin {
	return CPUEntitlementPlugin{}
}

func (p CPUEntitlementPlugin) Run(cli plugin.CliConnection, args []string) {
	if args[0] == "CLI-MESSAGE-UNINSTALL" {
		os.Exit(0)
	}

	traceLogger := trace.NewLogger(os.Stdout, true, os.Getenv("CF_TRACE"), "")
	ui := terminal.NewUI(os.Stdin, os.Stdout, terminal.NewTeePrinter(os.Stdout), traceLogger)

	if len(args) != 2 {
		ui.Failed("Usage: cf cpu-entitlement <APP_NAME>")
		os.Exit(1)
	}

	ui.Warn("Note: This feature is experimental.")

	dopplerURL, err := cli.DopplerEndpoint()
	if err != nil {
		ui.Failed(err.Error())
		os.Exit(1)
	}

	logCacheURL, err := buildLogCacheURL(dopplerURL)
	if err != nil {
		ui.Failed(err.Error())
		os.Exit(1)
	}

	infoGetter := metadata.NewInfoGetter(cli)
	tokenGetter := token.NewGetter(cli.AccessToken)
	metricsFetcher := metrics.NewFetcher(logCacheURL, tokenGetter)
	metricsCalculator := calculator.New()
	display := output.NewTerminalDisplay(ui)
	metricsRenderer := output.NewRenderer(display)

	appName := args[1]
	runner := NewRunner(infoGetter, metricsFetcher, metricsCalculator, metricsRenderer)
	res := runner.Run(appName, time.Now().Add(-Month), time.Now())
	if res.IsFailure {
		if res.ErrorMessage != "" {
			ui.Failed(res.ErrorMessage)
		}

		if res.WarningMessage != "" {
			ui.Warn(res.WarningMessage)
		}

		os.Exit(1)
	}
}

func (p CPUEntitlementPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "CPUEntitlementPlugin",
		Version: plugin.VersionType{
			Major: 0,
			Minor: 0,
			Build: 2,
		},
		Commands: []plugin.Command{
			{
				Name:     "cpu-entitlement",
				Alias:    "cpu",
				HelpText: "See cpu usage per app",
				UsageDetails: plugin.Usage{
					Usage: "cf cpu-entitlement APP_NAME",
				},
			},
		},
	}
}

func buildLogCacheURL(dopplerURL string) (string, error) {
	logStreamURL, err := url.Parse(dopplerURL)
	if err != nil {
		return "", err
	}

	regex, err := regexp.Compile("doppler(\\S+):443")
	if err != nil {
		return "", err
	}

	match := regex.FindStringSubmatch(logStreamURL.Host)

	if len(match) != 2 {
		return "", errors.New("Unable to parse log-stream endpoint from doppler URL")
	}

	logStreamURL.Scheme = "http"
	logStreamURL.Host = "log-cache" + match[1]

	return logStreamURL.String(), nil
}
