package integration_test

import (
	"strings"

	fetcherorg "code.cloudfoundry.org/cpu-entitlement-plugin/fetchers/org"
	. "code.cloudfoundry.org/cpu-entitlement-plugin/test_utils"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Fetcher", func() {
	var (
		org string
		uid string

		logCacheFetcher fetcherorg.LogCacheFetcher
		getToken        func() (string, error)
	)

	getUsages := func(appName string) func() []float64 {
		appGuid := getCmdOutput("cf", "app", appName, "--guid")
		return func() []float64 {
			usages, err := logCacheFetcher.FetchInstanceEntitlementUsages(appGuid)
			Expect(err).NotTo(HaveOccurred())
			return usages
		}
	}

	BeforeEach(func() {
		uid = uuid.New().String()
		org = "org-" + uid
		space := "space-" + uid

		Expect(Cmd("cf", "create-org", org).WithTimeout("3s").Run()).To(gexec.Exit(0))
		Expect(Cmd("cf", "target", "-o", org).Run()).To(gexec.Exit(0))
		Expect(Cmd("cf", "create-space", space).WithTimeout("3s").Run()).To(gexec.Exit(0))
		Expect(Cmd("cf", "target", "-o", org, "-s", space).Run()).To(gexec.Exit(0))

		logCacheURL := strings.Replace(cfApi, "https://api.", "http://log-cache.", 1)
		getToken = func() (string, error) {
			return getCmdOutput("cf", "oauth-token"), nil
		}

		logCacheFetcher = fetcherorg.NewLogCacheFetcher(logCacheURL, getToken)
	})

	AfterEach(func() {
		Expect(Cmd("cf", "delete-org", "-f", org).WithTimeout("1m").Run()).To(gexec.Exit(0))
	})

	When("running multiple apps with various instance counts", func() {
		BeforeEach(func() {
			Expect(Cmd("cf", "push", "spinner-1-"+uid, "-i", "3").WithDir("../../spinner").WithTimeout("2m").Run()).To(gexec.Exit(0))
			Expect(Cmd("cf", "push", "spinner-2-"+uid).WithDir("../../spinner").WithTimeout("2m").Run()).To(gexec.Exit(0))
		})

		It("gets the usages of all instances for each app", func() {
			Eventually(getUsages("spinner-1-"+uid), "20s", "1s").Should(HaveLen(3))
			Eventually(getUsages("spinner-2-"+uid), "20s", "1s").Should(HaveLen(1))
		})
	})

	When("an app has no instances", func() {
		BeforeEach(func() {
			Expect(Cmd("cf", "push", "spinner-"+uid, "-i", "0").WithDir("../../spinner").WithTimeout("2m").Run()).To(gexec.Exit(0))
		})

		It("returns an empty list of usages", func() {
			Consistently(getUsages("spinner-"+uid), "20s", "1s").Should(BeEmpty())
		})
	})

	When("the log-cache URL is not correct", func() {
		BeforeEach(func() {
			logCacheFetcher = fetcherorg.NewLogCacheFetcher("http://1.2.3:123", getToken)
		})

		It("returns an error about the url", func() {
			_, err := logCacheFetcher.FetchInstanceEntitlementUsages("anything")
			Expect(err).To(MatchError(ContainSubstring("dial")))
		})
	})
})

func getCmdOutput(cmd string, args ...string) string {
	return strings.TrimSpace(string(Cmd(cmd, args...).Run().Out.Contents()))
}
