package integration_test

import (
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"code.cloudfoundry.org/cpu-entitlement-plugin/httpclient"
	. "code.cloudfoundry.org/cpu-entitlement-plugin/test_utils"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	logcache "code.cloudfoundry.org/log-cache/pkg/client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(120 * time.Second)
	SetDefaultEventuallyPollingInterval(5 * time.Second)
	RunSpecs(t, "Integration Suite")
}

var (
	cfApi                string
	logEmitterHttpClient *http.Client
	logCacheClient       *logcache.Client
	getToken             func() (string, error)
	logger               lager.Logger
)

var _ = BeforeSuite(func() {
	cfApi = GetEnv("CF_API")
	cfUsername := GetEnv("CF_USERNAME")
	cfPassword := GetEnv("CF_PASSWORD")

	Expect(Cmd("cf", "api", cfApi, "--skip-ssl-validation").Run()).To(gexec.Exit(0))
	Expect(Cmd("cf", "login", "-u", cfUsername, "-p", cfPassword).Run()).To(gexec.Exit(0))

	logEmitterHttpClient = createInsecureHttpClient()
	Eventually(pingTestLogEmitter).Should(BeTrue())

	logCacheURL := getLogCacheURL()
	getToken = func() (string, error) {
		return getCmdOutput("cf", "oauth-token"), nil
	}
	logCacheClient = logcache.NewClient(
		logCacheURL,
		logcache.WithHTTPClient(httpclient.NewAuthClient(getToken)),
	)
	logger = lagertest.NewTestLogger("cumulative-usage-fetcher-test")
})

func createInsecureHttpClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig.InsecureSkipVerify = true
	return &http.Client{Transport: transport}
}

func GetEnv(varName string) string {
	value := os.Getenv(varName)
	ExpectWithOffset(1, value).NotTo(BeEmpty())
	return value
}

func pingTestLogEmitter() bool {
	response, err := logEmitterHttpClient.Get(getTestLogEmitterURL())
	if err != nil {
		return false
	}
	defer response.Body.Close()
	return response.StatusCode == http.StatusOK
}

func getTestLogEmitterURL() string {
	return strings.Replace(cfApi, "api.", "test-log-emitter.", 1)
}

func getLogCacheURL() string {
	return strings.Replace(cfApi, "https://api.", "http://log-cache.", 1)
}
