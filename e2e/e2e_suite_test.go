package e2e_test

import (
	"os"
	"testing"
	"time"

	. "code.cloudfoundry.org/cpu-entitlement-plugin/test_utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2e Suite")
}

var (
	cfApi string
)

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(5 * time.Second)
	Expect(Cmd("make", "install").WithDir("..").WithTimeout("30s").Run()).To(gexec.Exit(0))

	cfApi = GetEnv("CF_API")

	Expect(Cmd("cf", "api", cfApi, "--skip-ssl-validation").Run()).To(gexec.Exit(0))

	cfUsername := GetEnv("CF_USERNAME")
	cfPassword := GetEnv("CF_PASSWORD")

	Expect(Cmd("cf", "login", "-u", cfUsername, "-p", cfPassword).Run()).To(gexec.Exit(0))
})

func GetEnv(varName string) string {
	value := os.Getenv(varName)
	Expect(value).NotTo(BeEmpty())
	return value
}
