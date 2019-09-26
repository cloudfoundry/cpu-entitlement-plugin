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
	SetDefaultEventuallyTimeout(120 * time.Second)
	SetDefaultEventuallyPollingInterval(5 * time.Second)
	RunSpecs(t, "E2e Suite")
}

var _ = BeforeSuite(func() {
	Expect(Cmd("make", "install").WithDir("..").WithTimeout("30s").Run()).To(gexec.Exit(0))

})

func GetEnv(varName string) string {
	value := os.Getenv(varName)
	Expect(value).NotTo(BeEmpty())
	return value
}
