package plugin // import "code.cloudfoundry.org/cpu-entitlement-plugin/plugin"

import (
	"errors"
	"net/url"
	"os"
	"regexp"

	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/cf/trace"
	"code.cloudfoundry.org/cli/plugin"
	"code.cloudfoundry.org/cpu-entitlement-plugin/metadata"
	"code.cloudfoundry.org/cpu-entitlement-plugin/metricfetcher"
	"code.cloudfoundry.org/cpu-entitlement-plugin/output"
	"code.cloudfoundry.org/cpu-entitlement-plugin/token"
	"github.com/fatih/color"
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
		ui.Failed("Usage: `cf cpu-entitlement APP_NAME`")
		os.Exit(1)
	}

	appName := args[1]

	info, err := metadata.GetCFAppInfo(cli, appName)
	if err != nil {
		ui.Failed(err.Error())
		os.Exit(1)
	}

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

	tokenGetter := token.NewTokenGetter(cli.AccessToken)
	metricFetcher := metricfetcher.New(logCacheURL, tokenGetter)

	usageMetrics, err := metricFetcher.FetchLatest(info.App.Guid, info.App.InstanceCount)
	if err != nil {
		ui.Failed(err.Error())
		ui.Warn(bold("Your Cloud Foundry may not have enabled the CPU Entitlements feature. Please consult your operator."))
		os.Exit(1)
	}

	ui.Warn("Note: This feature is experimental.")

	metricsRenderer := output.NewRenderer(ui)
	metricsRenderer.ShowMetrics(info, usageMetrics)
}

func bold(message string) string {
	return terminal.Colorize(message, color.Bold)
}

func buildLogCacheURL(dopplerURL string) (string, error) {
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
	logStreamURL.Host = "log-cache" + match[1]

	return logStreamURL.String(), nil
}
