package e2e_test

import (
	"crypto/tls"
	"net/http"
	"strings"

	. "code.cloudfoundry.org/cpu-entitlement-plugin/test_utils"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("cpu-plugins", func() {
	var (
		org   string
		space string
		uid   string
	)

	BeforeEach(func() {
		uid = uuid.New().String()
		org = "org-" + uid
		space = "space-" + uid

		Expect(Cmd("cf", "create-org", org).WithTimeout("3s").Run()).To(gexec.Exit(0))
		Expect(Cmd("cf", "target", "-o", org).Run()).To(gexec.Exit(0))
		Expect(Cmd("cf", "create-space", space).WithTimeout("3s").Run()).To(gexec.Exit(0))
		Expect(Cmd("cf", "target", "-o", org, "-s", space).Run()).To(gexec.Exit(0))
	})

	AfterEach(func() {
		Expect(Cmd("cf", "delete-org", "-f", org).WithTimeout("1m").Run()).To(gexec.Exit(0))
	})

	Describe("cpu-entitlement-plugin", func() {
		var (
			appName string
		)

		BeforeEach(func() {
			appName = "spinner-" + uid
			Expect(Cmd("cf", "push", appName).WithDir("../../spinner").WithTimeout("2m").Run()).To(gexec.Exit(0))
		})

		It("prints the application entitlement info", func() {
			Eventually(Cmd("cf", "cpu-entitlement", appName).Run, "20s", "1s").Should(SatisfyAll(
				gbytes.Say("Showing CPU usage against entitlement for app %s in org %s / space %s as %s...", appName, org, space, cfUsername),
				gbytes.Say("avg usage"),
				gbytes.Say("#0"),
			))
		})

	})

	Describe("cpu-overentitlement-instances-plugin", func() {
		Describe("with an app", func() {
			var (
				overEntitlementApp  string
				appURL              string
				underEntitlementApp string
			)

			BeforeEach(func() {
				overEntitlementApp = "spinner-1-" + uid
				Expect(Cmd("cf", "push", overEntitlementApp).WithDir("../../spinner").WithTimeout("2m").Run()).To(gexec.Exit(0))
				appURL = strings.Replace(cfApi, "api.", overEntitlementApp+".", 1)

				underEntitlementApp = "spinner-2-" + uid
				Expect(Cmd("cf", "push", underEntitlementApp).WithDir("../../spinner").WithTimeout("2m").Run()).To(gexec.Exit(0))

				httpGet(appURL + "/spin")
			})

			It("prints the application over entitlement", func() {
				Eventually(Cmd("cf", "over-entitlement-instances").Run, "20s", "1s").Should(SatisfyAll(
					gbytes.Say("Showing over-entitlement apps in org %s as %s...", org, cfUsername),
					gbytes.Say("space *app"),
					gbytes.Say("%s *%s", space, overEntitlementApp),
				))
				Eventually(Cmd("cf", "over-entitlement-instances").Run).ShouldNot(gbytes.Say(underEntitlementApp))
			})
		})

		It("prints a no apps over messages if no apps over entitlement", func() {
			Consistently(Cmd("cf", "over-entitlement-instances").Run).Should(gbytes.Say("No apps over entitlement"))
		})
	})

	// TODO:test multiple spaces!
})

func httpGet(url string) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	resp, err := http.Get(url)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	ExpectWithOffset(1, resp.StatusCode).To(Equal(http.StatusOK))
}
