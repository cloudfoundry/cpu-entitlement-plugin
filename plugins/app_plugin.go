package plugins

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"time"

	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/cf/trace"
	"code.cloudfoundry.org/cli/plugin"
	"code.cloudfoundry.org/cpu-entitlement-plugin/cf"
	"code.cloudfoundry.org/cpu-entitlement-plugin/fetchers"
	"code.cloudfoundry.org/cpu-entitlement-plugin/httpclient"
	"code.cloudfoundry.org/cpu-entitlement-plugin/output"
	"code.cloudfoundry.org/cpu-entitlement-plugin/reporter"
	logcache "code.cloudfoundry.org/log-cache/pkg/client"
)

const month time.Duration = 31 * 24 * time.Hour

type CPUEntitlementPlugin struct{}

func NewCPUEntitlementPlugin() CPUEntitlementPlugin {
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

	logCacheURL, err := getLogCacheURL(cli)
	if err != nil {
		ui.Failed(err.Error())
		os.Exit(1)
	}

	sslIsDisabled, err := cli.IsSSLDisabled()
	if err != nil {
		ui.Failed(err.Error())
		os.Exit(1)
	}
	cfClient := cf.NewClient(cli, fetchers.NewProcessInstanceIDFetcher(createLogClient(logCacheURL, cli.AccessToken, sslIsDisabled)))
	historicalUsageFetcher := fetchers.NewHistoricalUsageFetcher(
		createLogClient(logCacheURL, cli.AccessToken, sslIsDisabled),
		time.Now().Add(-month),
		time.Now(),
	)
	currentUsageFetcher := fetchers.NewCurrentUsageFetcher(
		createLogClient(logCacheURL, cli.AccessToken, sslIsDisabled),
		time.Now().Add(-1*time.Minute), time.Now(),
	)
	metricsReporter := reporter.NewAppReporter(cfClient, historicalUsageFetcher, currentUsageFetcher)
	display := output.NewTerminalDisplay(ui)
	metricsRenderer := output.NewAppRenderer(display)

	appName := args[1]
	runner := NewAppRunner(metricsReporter, metricsRenderer)
	res := runner.Run(appName)
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

func getLogCacheURL(cli plugin.CliConnection) (string, error) {
	hasAPISet, err := cli.HasAPIEndpoint()
	if err != nil {
		return "", err
	}
	if !hasAPISet {
		return "", errors.New("No API endpoint set. Use 'cf login' or 'cf api' to target an endpoint.")
	}
	apiURL, err := cli.ApiEndpoint()
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`(https?://)[^.]+(\..*)`)
	match := re.FindStringSubmatch(apiURL)
	if len(match) != 3 {
		return "", fmt.Errorf("Unable to parse CF_API to get log-cache endpoint: %s", apiURL)
	}

	return match[1] + "log-cache" + match[2], nil

}

func createLogClient(logCacheURL string, accessTokenFunc func() (string, error), skipSSLValidation bool) *logcache.Client {
	httpClient := httpclient.NewAuthClient(accessTokenFunc)
	if skipSSLValidation {
		httpClient.SkipSSLValidation()
	}
	return logcache.NewClient(
		logCacheURL,
		logcache.WithHTTPClient(httpClient),
	)
}
