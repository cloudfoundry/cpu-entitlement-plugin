package e2e_test

import (
	"strings"
	"time"

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

		Expect(Cmd("cf", "create-org", org).Run()).To(gexec.Exit(0))
		Expect(Cmd("cf", "target", "-o", org).Run()).To(gexec.Exit(0))
		Expect(Cmd("cf", "create-space", space).Run()).To(gexec.Exit(0))
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
			Eventually(Cmd("cf", "cpu-entitlement", appName).Run).Should(SatisfyAll(
				gbytes.Say("Showing CPU usage against entitlement for app %s in org %s / space %s as", appName, org, space),
				gbytes.Say("avg usage"),
				gbytes.Say("#0"),
			))
		})

	})

	Describe("cpu-overentitlement-instances-plugin", func() {
		Describe("with an app", func() {
			var (
				appName string
				appURL  string
			)

			BeforeEach(func() {
				appName = "spinner-" + uid
				Expect(Cmd("cf", "push", appName).WithDir("../../spinner").WithTimeout("2m").Run()).To(gexec.Exit(0))
				appURL = strings.Replace(cfApi, "api.", appName+".", 1)
			})

			It("prints the list of apps that are over entitlement", func() {
				SpinFor(appURL, time.Second)
				Eventually(Cmd("cf", "over-entitlement-instances").Run).Should(gbytes.Say(appName))
			})
		})

		It("prints a no apps over messages if no apps over entitlement", func() {
			Consistently(Cmd("cf", "over-entitlement-instances").Run).Should(gbytes.Say("No apps over entitlement"))
		})
	})
})
