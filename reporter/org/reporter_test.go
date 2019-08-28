package org_test

import (
	"errors"

	"code.cloudfoundry.org/cpu-entitlement-plugin/cf"
	"code.cloudfoundry.org/cpu-entitlement-plugin/reporter/org"
	"code.cloudfoundry.org/cpu-entitlement-plugin/reporter/org/orgfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Reporter", func() {
	var (
		reporter           org.Reporter
		fakeCfClient       *orgfakes.FakeCloudFoundryClient
		fakeMetricsFetcher *orgfakes.FakeMetricsFetcher
	)

	BeforeEach(func() {
		fakeCfClient = new(orgfakes.FakeCloudFoundryClient)
		fakeMetricsFetcher = new(orgfakes.FakeMetricsFetcher)

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

		reporter = org.New(fakeCfClient, fakeMetricsFetcher)
	})

	Describe("OverEntitlementInstances", func() {
		var (
			report org.Report
			err    error
		)

		JustBeforeEach(func() {
			report, err = reporter.OverEntitlementInstances()
		})

		It("succeeds", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns all instances that are over entitlement", func() {
			Expect(report).To(Equal(org.Report{
				SpaceReports: []org.SpaceReport{
					org.SpaceReport{
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
	})
})
