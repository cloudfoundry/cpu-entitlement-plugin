package plugin // import "code.cloudfoundry.org/cpu-entitlement-plugin/plugin"

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/cf/trace"
	"code.cloudfoundry.org/cli/plugin"
	"code.cloudfoundry.org/cpu-entitlement-plugin/fetchers"
	"code.cloudfoundry.org/cpu-entitlement-plugin/metadata"
	"code.cloudfoundry.org/cpu-entitlement-plugin/output"
	"code.cloudfoundry.org/cpu-entitlement-plugin/reporter"
	"code.cloudfoundry.org/cpu-entitlement-plugin/token"
	logcache "code.cloudfoundry.org/log-cache/pkg/client"
)

const month time.Duration = 31 * 24 * time.Hour

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

	logCacheURL, err := getLogCacheURL(cli)
	if err != nil {
		ui.Failed(err.Error())
		os.Exit(1)
	}

	infoGetter := metadata.NewInfoGetter(cli)
	historicalUsageFetcher := fetchers.NewHistoricalUsageFetcher(
		createLogClient(logCacheURL, cli.AccessToken),
		time.Now().Add(-month),
		time.Now(),
	)
	currentUsageFetcher := fetchers.NewCurrentUsageFetcher(
		createLogClient(logCacheURL, cli.AccessToken),
	)
	metricsReporter := reporter.New(historicalUsageFetcher, currentUsageFetcher)
	display := output.NewTerminalDisplay(ui)
	metricsRenderer := output.NewRenderer(display)

	appName := args[1]
	runner := NewRunner(infoGetter, metricsReporter, metricsRenderer)
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
	dopplerURL, err := cli.DopplerEndpoint()
	if err != nil {
		return "", err
	}

	return buildLogCacheURL(dopplerURL)
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

func createLogClient(logCacheURL string, accessTokenFunc func() (string, error)) *logcache.Client {
	return logcache.NewClient(
		logCacheURL,
		logcache.WithHTTPClient(authenticatedBy(token.NewGetter(accessTokenFunc))),
	)
}

func authenticatedBy(tokenGetter *token.Getter) *authClient {
	return &authClient{tokenGetter: tokenGetter}
}

type authClient struct {
	tokenGetter *token.Getter
}

func (a *authClient) Do(req *http.Request) (*http.Response, error) {
	t, err := a.tokenGetter.Token()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", t)
	return http.DefaultClient.Do(req)
}
