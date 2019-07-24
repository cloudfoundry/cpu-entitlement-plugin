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
	"code.cloudfoundry.org/cpu-entitlement-plugin/result"
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
	tokenGetter := token.NewTokenGetter(cli.AccessToken)
	metricFetcher := metricfetcher.New(logCacheURL, tokenGetter)
	display := output.NewTerminalDisplay(ui)
	metricsRenderer := output.NewRenderer(display)

	runner := NewRunner(infoGetter, metricFetcher, metricsRenderer)
	res := runner.Run(args)
	if res.IsFailure() {
		res.WriteTo(ui)
		os.Exit(1)
	}
}

type Runner struct {
	infoGetter      metadata.InfoGetter
	metricFetcher   metricfetcher.CachedUsageMetricFetcher
	metricsRenderer output.Renderer
}

func NewRunner(infoGetter metadata.InfoGetter, metricFetcher metricfetcher.CachedUsageMetricFetcher, metricsRenderer output.Renderer) Runner {
	return Runner{infoGetter: infoGetter, metricFetcher: metricFetcher, metricsRenderer: metricsRenderer}
}

func (r Runner) Run(args []string) result.Result {
	if len(args) != 2 {
		return result.Failure("Usage: `cf cpu-entitlement APP_NAME`")
	}

	appName := args[1]

	info, err := r.infoGetter.GetCFAppInfo(appName)
	if err != nil {
		return result.FailureFromError(err)
	}

	usageMetrics, err := r.metricFetcher.FetchLatest(info.App.Guid, info.App.InstanceCount)
	if err != nil {
		return result.FailureFromError(err).WithWarning(bold("Your Cloud Foundry may not have enabled the CPU Entitlements feature. Please consult your operator."))
	}

	err = r.metricsRenderer.ShowMetrics(info, usageMetrics)
	if err != nil {
		return result.FailureFromError(err)
	}

	return result.Success()
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
