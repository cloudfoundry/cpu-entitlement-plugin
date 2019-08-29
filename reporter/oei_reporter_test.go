package reporter_test

import (
	"errors"

	"code.cloudfoundry.org/cpu-entitlement-plugin/cf"
	"code.cloudfoundry.org/cpu-entitlement-plugin/reporter"
	"code.cloudfoundry.org/cpu-entitlement-plugin/reporter/reporterfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Over-entitlement Instances Reporter", func() {
	var (
		oeiReporter        reporter.OverEntitlementInstances
		fakeCfClient       *reporterfakes.FakeCloudFoundryClient
		fakeMetricsFetcher *reporterfakes.FakeMetricsFetcher
	)

	BeforeEach(func() {
		fakeCfClient = new(reporterfakes.FakeCloudFoundryClient)
		fakeMetricsFetcher = new(reporterfakes.FakeMetricsFetcher)

		fakeCfClient.GetCurrentOrgReturns("org", nil)
		fakeCfClient.UsernameReturns("user", nil)
		fakeCfClient.GetSpacesReturns([]cf.Space{
			{
				Name: "space1",
				Applications: []cf.Application{
					{Name: "app1", Guid: "space1-app1-guid"},
					{Name: "app2", Guid: "space1-app2-guid"},
				},
			},
			{
				Name: "space2",
				Applications: []cf.Application{
					{Name: "app1", Guid: "space2-app1-guid"},
				},
			},
		}, nil)

		fakeMetricsFetcher.FetchInstanceEntitlementUsagesStub = func(appGuid string) ([]float64, error) {
			switch appGuid {
			case "space1-app1-guid":
				return []float64{1.5, 0.5}, nil
			case "space1-app2-guid":
				return []float64{0.3}, nil
			case "space2-app1-guid":
				return []float64{0.2}, nil
			}

			return nil, nil
		}

		oeiReporter = reporter.NewOverEntitlementInstances(fakeCfClient, fakeMetricsFetcher)
	})

	Describe("OverEntitlementInstances", func() {
		var (
			report reporter.OEIReport
			err    error
		)

		JustBeforeEach(func() {
			report, err = oeiReporter.OverEntitlementInstances()
		})

		It("succeeds", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns all instances that are over entitlement", func() {
			Expect(report).To(Equal(reporter.OEIReport{
				Org:      "org",
				Username: "user",
				SpaceReports: []reporter.SpaceReport{
					reporter.SpaceReport{
						SpaceName: "space1",
						Apps: []string{
							"app1",
						},
					},
				},
			}))
		})

		When("fetching the list of apps fails", func() {
			BeforeEach(func() {
				fakeCfClient.GetSpacesReturns(nil, errors.New("get-space-error"))
			})

			It("returns the error", func() {
				Expect(err).To(MatchError("get-space-error"))
			})
		})

		When("getting the entitlement usage for an app fails", func() {
			BeforeEach(func() {
				fakeMetricsFetcher.FetchInstanceEntitlementUsagesReturns(nil, errors.New("fetch-error"))
			})

			It("returns the error", func() {
				Expect(err).To(MatchError("fetch-error"))
			})
		})

		When("getting the current org fails", func() {
			BeforeEach(func() {
				fakeCfClient.GetCurrentOrgReturns("", errors.New("get-org-error"))
			})

			It("returns the error", func() {
				Expect(err).To(MatchError("get-org-error"))
			})
		})

		When("getting the username fails", func() {
			BeforeEach(func() {
				fakeCfClient.UsernameReturns("", errors.New("get-user-error"))
			})

			It("returns the error", func() {
				Expect(err).To(MatchError("get-user-error"))
			})
		})
	})
})
