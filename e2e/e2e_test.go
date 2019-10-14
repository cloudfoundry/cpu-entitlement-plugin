package e2e_test

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"code.cloudfoundry.org/cpu-entitlement-plugin/output"
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

	cfLogin := func() {
		Expect(Cmd("cf", "api", "--unset").Run()).To(gexec.Exit(0))
		getCFDetails()
		Expect(Cmd("cf", "api", cfApi, "--skip-ssl-validation").Run()).To(gexec.Exit(0))
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

	getAppUrl := func(appName string) string {
		return strings.Replace(cfApi, "api.", appName+".", 1)
	}

	Describe("using --skip-ssl-validation", func() {
		BeforeEach(func() {
			cfLogin()
			createAndTargetOrgAndSpace()
		})

		AfterEach(func() {
			Expect(Cmd("cf", "delete-org", "-f", org).WithTimeout("1m").Run()).To(gexec.Exit(0))
		})

		Describe("cpu-entitlement-plugin", func() {
			var (
				appName   string
				appStart  time.Time
				instances int
			)

			BeforeEach(func() {
				instances = 1
			})

			JustBeforeEach(func() {
				appName = "spinner-" + uid
				appStart = time.Now()
				PushSpinner(appName, instances)
			})

			It("prints the application entitlement info", func() {
				Eventually(Cmd("cf", "cpu-entitlement", appName).Run).Should(SatisfyAll(
					gbytes.Say("Showing CPU usage against entitlement for app %s in org %s / space %s as %s...", appName, org, space, cfUsername),
					gbytes.Say("avg usage"),
					gbytes.Say("#0"),
				))
			})

			Describe("spikes", func() {
				var waitgroup sync.WaitGroup

				getAvgUsageFunc := func(idx int) func() (float64, error) {
					return func() (float64, error) {
						return getAvgUsage(appName, idx)
					}
				}

				waitForSpike := func(instanceId int) {
					defer GinkgoRecover()
					defer waitgroup.Done()

					Eventually(getAvgUsageFunc(instanceId)).Should(BeNumerically(">", 100))
					Eventually(getAvgUsageFunc(instanceId)).Should(BeNumerically("<", 100))
				}

				BeforeEach(func() {
					instances = 2
				})

				JustBeforeEach(func() {
					Spin(getAppUrl(appName), 2)
					Spin(getAppUrl(appName), 2)

					waitgroup.Add(2)

					go waitForSpike(0)
					go waitForSpike(1)

					waitgroup.Wait()
				})

				It("shows last spike", func() {
					for i := 0; i < 2; i++ {
						start, end := getLastSpike(appName, i)
						Expect(start).To(BeTemporally(">", appStart), fmt.Sprintf("instance %d", i))
						Expect(end).To(BeTemporally("<", time.Now()), fmt.Sprintf("instance %d", i))
					}
				})
			})
		})

		Describe("cpu-overentitlement-instances-plugin", func() {
			Describe("with an app", func() {
				var (
					overEntitlementApp  string
					underEntitlementApp string
				)

				BeforeEach(func() {
					var wg sync.WaitGroup
					wg.Add(2)

					overEntitlementApp = "overentitled-app-" + uid
					underEntitlementApp = "underentitled-app-" + uid

					go func() {
						defer GinkgoRecover()
						defer wg.Done()
						PushSpinner(overEntitlementApp, 1)
						Spin(getAppUrl(overEntitlementApp), 0)
					}()

					go func() {
						defer GinkgoRecover()
						defer wg.Done()
						PushSpinner(underEntitlementApp, 1)
					}()

					wg.Wait()
				})

				AfterEach(func() {
					Unspin(getAppUrl(overEntitlementApp))
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
						anotherSpace string
						anotherApp   string
					)

					BeforeEach(func() {
						anotherApp = "anotherspinner-1-" + uid
						anotherSpace = "anotherspace" + uid
						Expect(Cmd("cf", "create-space", anotherSpace).Run()).To(gexec.Exit(0))
						Expect(Cmd("cf", "target", "-o", org, "-s", anotherSpace).Run()).To(gexec.Exit(0))
						PushSpinner(anotherApp, 1)
						Spin(getAppUrl(anotherApp), 0)
					})

					AfterEach(func() {
						Unspin(getAppUrl(anotherApp))
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

	Describe("providing a CA cert rather than using --skip-ssl-validation", func() {
		var (
			certFile string
		)

		cfLogin := func() {
			Expect(Cmd("cf", "api", "--unset").WithEnv("SSL_CERT_FILE", certFile).Run()).To(gexec.Exit(0))
			getCFDetails()
			Expect(Cmd("cf", "api", cfApi).WithEnv("SSL_CERT_FILE", certFile).Run()).To(gexec.Exit(0))
			cfPassword := GetEnv("CF_PASSWORD")
			Expect(Cmd("cf", "auth", cfUsername, cfPassword).WithEnv("SSL_CERT_FILE", certFile).Run()).To(gexec.Exit(0))
		}

		createAndTargetOrgAndSpace := func() {
			uid = uuid.New().String()
			org = "org-" + uid
			space = "space-" + uid

			Expect(Cmd("cf", "create-org", org).WithEnv("SSL_CERT_FILE", certFile).Run()).To(gexec.Exit(0))
			Expect(Cmd("cf", "target", "-o", org).WithEnv("SSL_CERT_FILE", certFile).Run()).To(gexec.Exit(0))
			Expect(Cmd("cf", "create-space", space).WithEnv("SSL_CERT_FILE", certFile).Run()).To(gexec.Exit(0))
			Expect(Cmd("cf", "target", "-o", org, "-s", space).WithEnv("SSL_CERT_FILE", certFile).Run()).To(gexec.Exit(0))
		}

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
			cfLogin()
			createAndTargetOrgAndSpace()
		})

		AfterEach(func() {
			Expect(Cmd("cf", "delete-org", "-f", org).WithEnv("SSL_CERT_FILE", certFile).WithTimeout("1m").Run()).To(gexec.Exit(0))
		})

		It("should successfully run entitlement plugin when SSL_CERT_FILE is set to a valid cert file", func() {
			appName := "spinner-" + uid
			PushSpinnerWithCert(appName, 1, certFile)

			Expect(Cmd("cf", "cpu-entitlement", appName).WithEnv("SSL_CERT_FILE", certFile).Run()).To(SatisfyAll(
				gexec.Exit(0),
				gbytes.Say(appName),
			))
		})

		It("should successfully run oei plugin when SSL_CERT_FILE is set to a valid cert file", func() {
			Expect(Cmd("cf", "over-entitlement-instances").WithEnv("SSL_CERT_FILE", certFile).Run()).To(SatisfyAll(
				gexec.Exit(0),
				gbytes.Say(org),
			))
		})

		When("SSL_CERT_FILE not set", func() {
			It("should exit entitlement plugin with non-zero status", func() {
				Expect(Cmd("cf", "cpu-entitlement", "myapp").Run()).To(SatisfyAll(
					gexec.Exit(1),
					gbytes.Say("unknown authority"),
				))
			})
		})
	})
})

func getAvgUsage(appName string, instanceIndex int) (float64, error) {
	cmd := fmt.Sprintf(`cf cpu %s | grep "^#%d" | awk '{ print $2 }' | tr -d '\n, ,%%'`, appName, instanceIndex)
	out, err := exec.Command("/bin/bash", "-c", cmd).CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "Error reading cpu usage: %s", string(out))

	return strconv.ParseFloat(string(out), 64)
}

func getLastSpike(appName string, instanceIndex int) (time.Time, time.Time) {
	out, err := exec.Command("/bin/bash", "-c",
		fmt.Sprintf(`cf cpu %s | grep "WARNING: Instance #%d was over entitlement"`, appName, instanceIndex)).CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "Error getting last spike: %s, instance:%d", string(out), instanceIndex)

	r := regexp.MustCompile(fmt.Sprintf("WARNING: Instance #%d was over entitlement from (.*) to (.*)", instanceIndex))
	matches := r.FindStringSubmatch(string(out))
	Expect(matches).To(HaveLen(3))

	spikeStart, err := time.Parse(output.DateFmt, matches[1])
	Expect(err).NotTo(HaveOccurred())
	spikeEnd, err := time.Parse(output.DateFmt, matches[2])
	Expect(err).NotTo(HaveOccurred())

	return spikeStart, spikeEnd
}
