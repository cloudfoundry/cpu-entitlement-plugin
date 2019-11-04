package integration_test

import (
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	. "code.cloudfoundry.org/cpu-entitlement-plugin/test_utils"
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
)

var _ = BeforeSuite(func() {
	cfApi = GetEnv("CF_API")

	Expect(Cmd("cf", "api", cfApi, "--skip-ssl-validation").Run()).To(gexec.Exit(0))

	cfUsername := GetEnv("CF_USERNAME")
	cfPassword := GetEnv("CF_PASSWORD")

	Expect(Cmd("cf", "login", "-u", cfUsername, "-p", cfPassword).Run()).To(gexec.Exit(0))

	logEmitterHttpClient = createInsecureHttpClient()
	Eventually(pingTestLogEmitter).Should(BeTrue())
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
