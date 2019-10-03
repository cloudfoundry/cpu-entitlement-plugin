package e2e_test

import (
	"io/ioutil"
	"os"
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
		cfApi      string
		cfUsername string
		org        string
		space      string
		uid        string
	)

	getCFDetails := func() {
		cfApi = GetEnv("CF_API")
		cfUsername = GetEnv("CF_USERNAME")
	}

	cfLogin := func(skipSSLValidation bool) {
		Expect(Cmd("cf", "api", "--unset").Run()).To(gexec.Exit(0))
		getCFDetails()
		if skipSSLValidation {
			Expect(Cmd("cf", "api", cfApi, "--skip-ssl-validation").Run()).To(gexec.Exit(0))
		} else {
			Expect(Cmd("cf", "api", cfApi).Run()).To(gexec.Exit(0))
		}
		cfPassword := GetEnv("CF_PASSWORD")
		Expect(Cmd("cf", "auth", cfUsername, cfPassword).Run()).To(gexec.Exit(0))
	}

	createAndTargetOrgAndSpace := func() {
		uid = uuid.New().String()
		org = "org-" + uid
		space = "space-" + uid

		Expect(Cmd("cf", "create-org", org).Run()).To(gexec.Exit(0))
		Expect(Cmd("cf", "target", "-o", org).Run()).To(gexec.Exit(0))
		Expect(Cmd("cf", "create-space", space).Run()).To(gexec.Exit(0))
		Expect(Cmd("cf", "target", "-o", org, "-s", space).Run()).To(gexec.Exit(0))
	}

	When("skipping SSL validation", func() {

		BeforeEach(func() {
			cfLogin(true)
			createAndTargetOrgAndSpace()
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

	When("providing a CA cert rather than using --skip-ssl-validation", func() {
		var (
			certFile string
		)

		writeCertFile := func() string {
			cert := GetEnv("ROUTER_CA_CERT")
			tmpFile, err := ioutil.TempFile("", "cert")
			Expect(err).NotTo(HaveOccurred())
			certFileName := tmpFile.Name()
			Expect(ioutil.WriteFile(certFileName, []byte(cert), 0400)).To(Succeed())
			return certFileName
		}

		BeforeEach(func() {
			certFile = writeCertFile()
			os.Setenv("SSL_CERT_FILE", certFile)

			cfLogin(false)
			createAndTargetOrgAndSpace()
		})

		AfterEach(func() {
			os.Setenv("SSL_CERT_FILE", certFile)
			Expect(Cmd("cf", "delete-org", "-f", org).WithTimeout("1m").Run()).To(gexec.Exit(0))
			os.Unsetenv("SSL_CERT_FILE")
		})

		It("should successfully run entitlement plugin when SSL_CERT_FILE is set to a valid cert file", func() {
			appName := "spinner-" + uid
			PushSpinner(appName, 1)

			Expect(Cmd("cf", "cpu-entitlement", appName).Run()).To(SatisfyAll(
				gexec.Exit(0),
				gbytes.Say(appName),
			))
		})

		It("should successfully run oei plugin when SSL_CERT_FILE is set to a valid cert file", func() {
			Expect(Cmd("cf", "over-entitlement-instances").Run()).To(SatisfyAll(
				gexec.Exit(0),
				gbytes.Say(org),
			))
		})

		It("should exit entitlement plugin with non-zero status if SSL_CERT_FILE not set and --skip-ssl-validation not passed", func() {
			os.Unsetenv("SSL_CERT_FILE")

			Expect(Cmd("cf", "cpu-entitlement", "myapp").Run()).To(SatisfyAll(
				gexec.Exit(1),
				gbytes.Say("unknown authority"),
			))
		})
	})
})
