package e2e_test

import (
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
			PushSpinner(appName, 1)
		})

		It("prints the application entitlement info", func() {
			Eventually(Cmd("cf", "cpu-entitlement", appName).Run).Should(SatisfyAll(
				gbytes.Say("Showing CPU usage against entitlement for app %s in org %s / space %s as %s...", appName, org, space, cfUsername),
				gbytes.Say("avg usage"),
				gbytes.Say("#0"),
			))
		})

	})

	Describe("cpu-overentitlement-instances-plugin", func() {
		Describe("with an app", func() {
			var (
				overEntitlementApp    string
				overEntitlementAppURL string
				underEntitlementApp   string
			)

			BeforeEach(func() {
				overEntitlementApp = "overentitled-app-" + uid
				PushSpinner(overEntitlementApp, 1)
				overEntitlementAppURL = strings.Replace(cfApi, "api.", overEntitlementApp+".", 1)
				Spin(overEntitlementAppURL)

				underEntitlementApp = "underentitled-app-" + uid
				PushSpinner(underEntitlementApp, 1)
			})

			AfterEach(func() {
				Unspin(overEntitlementAppURL)
			})

			It("prints the application over entitlement", func() {
				Eventually(Cmd("cf", "over-entitlement-instances").Run).Should(SatisfyAll(
					gbytes.Say("Showing over-entitlement apps in org %s as %s...", org, cfUsername),
					gbytes.Say("space *app"),
					gbytes.Say("%s *%s", space, overEntitlementApp),
				))
				Eventually(Cmd("cf", "over-entitlement-instances").Run).ShouldNot(gbytes.Say(underEntitlementApp))
			})

			When("over entitlement apps are in different spaces", func() {
				var (
					anotherSpace  string
					anotherApp    string
					anotherAppURL string
				)

				BeforeEach(func() {
					anotherApp = "anotherspinner-1-" + uid
					anotherSpace = "anotherspace" + uid
					anotherAppURL = strings.Replace(cfApi, "api.", anotherApp+".", 1)
					Expect(Cmd("cf", "create-space", anotherSpace).Run()).To(gexec.Exit(0))
					Expect(Cmd("cf", "target", "-o", org, "-s", anotherSpace).Run()).To(gexec.Exit(0))
					PushSpinner(anotherApp, 1)
					Spin(anotherAppURL)
				})

				AfterEach(func() {
					Unspin(anotherAppURL)
				})

				It("prints apps over entitlement from different spaces", func() {
					Eventually(Cmd("cf", "over-entitlement-instances").Run).Should(SatisfyAll(
						gbytes.Say("Showing over-entitlement apps in org %s as %s...", org, cfUsername),
						gbytes.Say("space *app"),
						gbytes.Say("%s *%s", anotherSpace, anotherApp),
						gbytes.Say("%s *%s", space, overEntitlementApp),
					))
				})
			})
		})

		It("prints a no apps over messages if no apps over entitlement", func() {
			Consistently(Cmd("cf", "over-entitlement-instances").Run).Should(gbytes.Say("No apps over entitlement"))
		})
	})
})
