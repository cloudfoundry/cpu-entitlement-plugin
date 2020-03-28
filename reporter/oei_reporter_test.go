package reporter_test

import (
	"errors"

	"code.cloudfoundry.org/cpu-entitlement-plugin/cf"
	"code.cloudfoundry.org/cpu-entitlement-plugin/fetchers"
	"code.cloudfoundry.org/cpu-entitlement-plugin/reporter"
	"code.cloudfoundry.org/cpu-entitlement-plugin/reporter/reporterfakes"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Over-entitlement Instances Reporter", func() {
	var (
		oeiReporter        reporter.OverEntitlementInstances
		fakeCfClient       *reporterfakes.FakeCloudFoundryClient
		fakeMetricsFetcher *reporterfakes.FakeMetricsFetcher
		report             reporter.OEIReport
		logger             lager.Logger
		err                error
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

		fakeMetricsFetcher.FetchInstanceDataStub = func(logger lager.Logger, appGuid string, appInstances map[int]cf.Instance) (map[int]interface{}, error) {
			switch appGuid {
			case "space1-app1-guid":
				return map[int]interface{}{
					0: fetchers.CumulativeInstanceData{Usage: 1.5},
					1: fetchers.CumulativeInstanceData{Usage: 0.5},
				}, nil
			case "space1-app2-guid":
				return map[int]interface{}{
					0: fetchers.CumulativeInstanceData{Usage: 0.3},
				}, nil
			case "space2-app1-guid":
				return map[int]interface{}{
					0: fetchers.CumulativeInstanceData{Usage: 0.2},
				}, nil
			}

			return nil, nil
		}

		logger = lagertest.NewTestLogger("oei-reporter-test")

		oeiReporter = reporter.NewOverEntitlementInstances(fakeCfClient, fakeMetricsFetcher)
	})

	JustBeforeEach(func() {
		report, err = oeiReporter.OverEntitlementInstances(logger)
	})

	It("succeeds", func() {
		Expect(err).NotTo(HaveOccurred())
	})

	It("logs start and end of function", func() {
		Expect(logger).To(gbytes.Say("oei-reporter.start"))
		Expect(logger).To(gbytes.Say("oei-reporter.end"))
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

		It("logs the error", func() {
			Expect(logger).To(SatisfyAll(
				gbytes.Say("failed-to-get-spaces"),
				gbytes.Say(`"log_level":2`),
				gbytes.Say("get-space-error"),
			))
		})
	})

	When("getting the entitlement usage for an app fails", func() {
		BeforeEach(func() {
			fakeMetricsFetcher.FetchInstanceDataReturns(nil, errors.New("fetch-error"))
		})

		It("returns the error", func() {
			Expect(err).To(MatchError("fetch-error"))
		})

		It("logs the error", func() {
			Expect(logger).To(SatisfyAll(
				gbytes.Say("failed-to-fetch-instance-metrics"),
				gbytes.Say(`"log_level":2`),
				gbytes.Say(`"app-guid":"space1-app1-guid"`),
				gbytes.Say("fetch-error"),
			))
		})
	})

	When("getting the current org fails", func() {
		BeforeEach(func() {
			fakeCfClient.GetCurrentOrgReturns("", errors.New("get-org-error"))
		})

		It("returns the error", func() {
			Expect(err).To(MatchError("get-org-error"))
		})

		It("logs the error", func() {
			Expect(logger).To(SatisfyAll(
				gbytes.Say("failed-to-get-current-org"),
				gbytes.Say(`"log_level":2`),
				gbytes.Say("get-org-error"),
			))
		})
	})

	When("getting the username fails", func() {
		BeforeEach(func() {
			fakeCfClient.UsernameReturns("", errors.New("get-user-error"))
		})

		It("returns the error", func() {
			Expect(err).To(MatchError("get-user-error"))
		})

		It("logs the error", func() {
			Expect(logger).To(SatisfyAll(
				gbytes.Say("failed-to-get-username"),
				gbytes.Say(`"log_level":2`),
				gbytes.Say("get-user-error"),
			))
		})
	})

	When("the fetcher returns the wrong type", func() {
		BeforeEach(func() {
			fakeMetricsFetcher.FetchInstanceDataStub = func(_ lager.Logger, appGuid string, appInstances map[int]cf.Instance) (map[int]interface{}, error) {
				switch appGuid {
				case "space1-app1-guid":
					return map[int]interface{}{
						0: "hello",
					}, nil
				case "space1-app2-guid":
					return map[int]interface{}{
						0: fetchers.CumulativeInstanceData{Usage: 1.3},
					}, nil
				}

				return nil, nil
			}
		})

		It("skips the instance with the wrong type", func() {
			Expect(len(report.SpaceReports)).To(Equal(1))
			Expect(len(report.SpaceReports[0].Apps)).To(Equal(1))
			Expect(report.SpaceReports[0].Apps[0]).To(Equal("app2"))
		})

		It("logs the wrong type", func() {
			Expect(logger).To(SatisfyAll(
				gbytes.Say("metrics-fetcher-returned-wrong-type"),
				gbytes.Say(`"log_level":1`),
				gbytes.Say(`"instance-data":"hello"`),
			))
		})
	})

	When("spaces are not sorted alphabetically", func() {
		BeforeEach(func() {
			fakeCfClient.GetSpacesReturns([]cf.Space{
				{
					Name: "space2",
					Applications: []cf.Application{
						{Name: "app1", Guid: "space2-app1-guid"},
					},
				},
				{
					Name: "space1",
					Applications: []cf.Application{
						{Name: "app1", Guid: "space1-app1-guid"},
					},
				},
			}, nil)
			fakeMetricsFetcher.FetchInstanceDataReturns(
				map[int]interface{}{
					0: fetchers.CumulativeInstanceData{Usage: 1.5},
				}, nil)
		})

		It("reports sorted spaces", func() {
			Expect(len(report.SpaceReports)).To(Equal(2))
			Expect(report.SpaceReports[0].SpaceName).To(Equal("space1"))
			Expect(report.SpaceReports[1].SpaceName).To(Equal("space2"))
		})
	})

	When("apps in a single space are not sorted alphabetically", func() {
		BeforeEach(func() {
			fakeCfClient.GetSpacesReturns([]cf.Space{
				{
					Name: "space1",
					Applications: []cf.Application{
						{Name: "app2", Guid: "space1-app2-guid"},
						{Name: "app1", Guid: "space1-app1-guid"},
					},
				},
			}, nil)
			fakeMetricsFetcher.FetchInstanceDataReturns(
				map[int]interface{}{
					0: fetchers.CumulativeInstanceData{Usage: 1.5},
				}, nil)
		})

		It("reports sorted apps in the report", func() {
			Expect(len(report.SpaceReports)).To(Equal(1))
			Expect(len(report.SpaceReports[0].Apps)).To(Equal(2))
			Expect(report.SpaceReports[0].Apps[0]).To(Equal("app1"))
			Expect(report.SpaceReports[0].Apps[1]).To(Equal("app2"))
		})
	})
})
