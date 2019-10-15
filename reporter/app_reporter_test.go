package reporter_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cpu-entitlement-plugin/cf"
	"code.cloudfoundry.org/cpu-entitlement-plugin/fetchers"
	"code.cloudfoundry.org/cpu-entitlement-plugin/reporter"
	"code.cloudfoundry.org/cpu-entitlement-plugin/reporter/reporterfakes"
)

var _ = Describe("Reporter", func() {
	var (
		cumulativeUsageFetcher *reporterfakes.FakeInstanceDataFetcher
		currentUsageFetcher    *reporterfakes.FakeInstanceDataFetcher
		lastSpikeFetcher       *reporterfakes.FakeInstanceDataFetcher
		cfClient               *reporterfakes.FakeAppReporterCloudFoundryClient
		instanceReporter       reporter.AppReporter
		reports                reporter.ApplicationReport
		err                    error
		appName                string
		appGuid                string
		appInstances           map[int]cf.Instance
	)

	BeforeEach(func() {
		appName = "foo"
		appGuid = "bar"

		cumulativeUsageFetcher = new(reporterfakes.FakeInstanceDataFetcher)
		currentUsageFetcher = new(reporterfakes.FakeInstanceDataFetcher)
		lastSpikeFetcher = new(reporterfakes.FakeInstanceDataFetcher)
		cfClient = new(reporterfakes.FakeAppReporterCloudFoundryClient)
		appInstances = map[int]cf.Instance{0: cf.Instance{InstanceID: 0}}
		cfClient.GetApplicationReturns(cf.Application{Name: appName, Guid: appGuid, Instances: appInstances}, nil)
		cfClient.GetCurrentOrgReturns("the-org", nil)
		cfClient.GetCurrentSpaceReturns("the-space", nil)
		cfClient.UsernameReturns("the-user", nil)

		currentUsageFetcher.FetchInstanceDataReturns(map[int]interface{}{
			0: fetchers.CurrentInstanceData{
				InstanceID: 0,
				Usage:      0.5,
			},
		}, nil)

		instanceReporter = reporter.NewAppReporter(cfClient, currentUsageFetcher, lastSpikeFetcher, cumulativeUsageFetcher)
	})

	JustBeforeEach(func() {
		reports, err = instanceReporter.CreateApplicationReport(appName)
	})

	Describe("Report", func() {
		BeforeEach(func() {
			cumulativeUsageFetcher.FetchInstanceDataReturns(map[int]interface{}{
				0: fetchers.CumulativeInstanceData{
					InstanceID: 0,
					Usage:      0.5,
				},
			}, nil)
			currentUsageFetcher.FetchInstanceDataReturns(map[int]interface{}{
				0: fetchers.CurrentInstanceData{
					InstanceID: 0,
					Usage:      1.5,
				},
			}, nil)
			lastSpikeFetcher.FetchInstanceDataReturns(map[int]interface{}{
				0: fetchers.LastSpikeInstanceData{
					InstanceID: 0,
					From:       time.Unix(5, 0),
					To:         time.Unix(10, 0),
				},
			}, nil)
		})

		It("reports the application name", func() {
			Expect(reports.ApplicationName).To(Equal(appName))
		})

		When("getting the application errors", func() {
			BeforeEach(func() {
				cfClient.GetApplicationReturns(cf.Application{}, errors.New("app error"))
			})

			It("returns the error", func() {
				Expect(err).To(MatchError("app error"))
			})
		})

		It("reports the org", func() {
			Expect(reports.Org).To(Equal("the-org"))
		})

		When("getting the org errors", func() {
			BeforeEach(func() {
				cfClient.GetCurrentOrgReturns("", errors.New("org error"))
			})

			It("returns the error", func() {
				Expect(err).To(MatchError("org error"))
			})
		})

		It("reports the space", func() {
			Expect(reports.Space).To(Equal("the-space"))
		})

		When("getting the space errors", func() {
			BeforeEach(func() {
				cfClient.GetCurrentSpaceReturns("", errors.New("space error"))
			})

			It("returns the error", func() {
				Expect(err).To(MatchError("space error"))
			})
		})

		It("reports the user", func() {
			Expect(reports.Username).To(Equal("the-user"))
		})

		When("getting the user errors", func() {
			BeforeEach(func() {
				cfClient.UsernameReturns("", errors.New("user error"))
			})

			It("returns the error", func() {
				Expect(err).To(MatchError("user error"))
			})
		})

		It("combines current usage, cumulative usage and last spike data", func() {
			Expect(len(reports.InstanceReports)).To(Equal(1))

			Expect(reports.InstanceReports[0].InstanceID).To(Equal(0))
			Expect(reports.InstanceReports[0].CumulativeUsage.Value).To(Equal(0.5))
			Expect(reports.InstanceReports[0].CurrentUsage.Value).To(Equal(1.5))
			Expect(reports.InstanceReports[0].LastSpike).To(Equal(reporter.LastSpike{From: time.Unix(5, 0), To: time.Unix(10, 0)}))
		})

		When("current usage data, historical usage data and last spike data cannot be matched by instance id", func() {
			BeforeEach(func() {
				cumulativeUsageFetcher.FetchInstanceDataReturns(map[int]interface{}{
					0: fetchers.CumulativeInstanceData{
						InstanceID: 0,
						Usage:      0.5,
					},
				}, nil)
				currentUsageFetcher.FetchInstanceDataReturns(map[int]interface{}{
					1: fetchers.CurrentInstanceData{
						InstanceID: 1,
						Usage:      1.5,
					},
				}, nil)
				lastSpikeFetcher.FetchInstanceDataReturns(map[int]interface{}{
					2: fetchers.LastSpikeInstanceData{
						InstanceID: 2,
						From:       time.Unix(5, 0),
						To:         time.Unix(10, 0),
					},
				}, nil)
			})

			It("combines cumulative cpu usage, current cpu usage and last spike data", func() {
				Expect(len(reports.InstanceReports)).To(Equal(3))

				Expect(reports.InstanceReports[0].InstanceID).To(Equal(0))
				Expect(reports.InstanceReports[0].CumulativeUsage.Value).To(Equal(0.5))
				Expect(reports.InstanceReports[0].CurrentUsage.Value).To(BeZero())
				Expect(reports.InstanceReports[0].LastSpike).To(Equal(reporter.LastSpike{}))

				Expect(reports.InstanceReports[1].InstanceID).To(Equal(1))
				Expect(reports.InstanceReports[1].CumulativeUsage.Value).To(BeZero())
				Expect(reports.InstanceReports[1].CurrentUsage.Value).To(Equal(1.5))
				Expect(reports.InstanceReports[1].LastSpike).To(Equal(reporter.LastSpike{}))

				Expect(reports.InstanceReports[2].InstanceID).To(Equal(2))
				Expect(reports.InstanceReports[2].CumulativeUsage.Value).To(BeZero())
				Expect(reports.InstanceReports[2].CurrentUsage.Value).To(BeZero())
				Expect(reports.InstanceReports[2].LastSpike).To(Equal(reporter.LastSpike{From: time.Unix(5, 0), To: time.Unix(10, 0)}))
			})
		})
	})

	Describe("Cumulative CPU usage", func() {
		BeforeEach(func() {
			cumulativeUsageFetcher.FetchInstanceDataReturns(map[int]interface{}{
				0: fetchers.CumulativeInstanceData{
					InstanceID: 0,
					Usage:      0.5,
				},
				1: fetchers.CumulativeInstanceData{
					InstanceID: 1,
					Usage:      0.7,
				},
			}, nil)
		})

		It("fetches the usage data correctly", func() {
			Expect(cumulativeUsageFetcher.FetchInstanceDataCallCount()).To(Equal(1))
			actualAppGuid, actualAppInstances := cumulativeUsageFetcher.FetchInstanceDataArgsForCall(0)
			Expect(actualAppGuid).To(Equal(appGuid))
			Expect(actualAppInstances).To(Equal(appInstances))
		})

		When("fetching the cumulative usage fails", func() {
			BeforeEach(func() {
				cumulativeUsageFetcher.FetchInstanceDataReturns(nil, errors.New("fetch-historical-error"))
			})

			It("returns the error", func() {
				Expect(err).To(MatchError("fetch-historical-error"))
			})
		})

		When("the fetcher does not return any data", func() {
			BeforeEach(func() {
				currentUsageFetcher.FetchInstanceDataReturns(map[int]interface{}{}, nil)
			})

			It("returns an UnsupportedCFDeploymentError", func() {
				Expect(err).To(MatchError(reporter.NewUnsupportedCFDeploymentError(appName)))
			})
		})

		When("the fetcher returns the wrong type of instance data", func() {
			BeforeEach(func() {
				cumulativeUsageFetcher.FetchInstanceDataReturns(map[int]interface{}{
					0: "hello",
					1: fetchers.CumulativeInstanceData{Usage: 0.1},
				}, nil)
			})

			It("skips the instance with the wrong type", func() {
				Expect(len(reports.InstanceReports)).To(Equal(2))
				Expect(reports.InstanceReports[0].CumulativeUsage).To(Equal(reporter.CumulativeUsage{}))
				Expect(reports.InstanceReports[1].CumulativeUsage).To(Equal(reporter.CumulativeUsage{Value: 0.1}))
			})
		})

		It("reports cumulative usage", func() {
			Expect(len(reports.InstanceReports)).To(Equal(2))

			Expect(reports.InstanceReports[0].InstanceID).To(Equal(0))
			Expect(reports.InstanceReports[0].CumulativeUsage.Value).To(Equal(0.5))

			Expect(reports.InstanceReports[1].InstanceID).To(Equal(1))
			Expect(reports.InstanceReports[1].CumulativeUsage.Value).To(Equal(0.7))
		})
	})

	Describe("Last spike", func() {
		It("fetches the spike data correctly", func() {
			Expect(lastSpikeFetcher.FetchInstanceDataCallCount()).To(Equal(1))
			actualAppGuid, actualAppInstances := cumulativeUsageFetcher.FetchInstanceDataArgsForCall(0)
			Expect(actualAppGuid).To(Equal(appGuid))
			Expect(actualAppInstances).To(Equal(appInstances))
		})

		When("fetching the last spike fails", func() {
			BeforeEach(func() {
				lastSpikeFetcher.FetchInstanceDataReturns(nil, errors.New("fetch-spike-error"))
			})

			It("returns the error", func() {
				Expect(err).To(MatchError("fetch-spike-error"))
			})
		})

		When("the fetcher returns the wrong type of instance data", func() {
			BeforeEach(func() {
				lastSpikeFetcher.FetchInstanceDataReturns(map[int]interface{}{
					0: "hello",
					1: fetchers.LastSpikeInstanceData{
						InstanceID: 1,
						From:       time.Unix(1, 0),
						To:         time.Unix(2, 0),
					},
				}, nil)
			})

			It("skips the instance with the wrong type", func() {
				Expect(reports.InstanceReports).To(HaveLen(2))
				Expect(reports.InstanceReports[0].LastSpike.From).To(Equal(time.Time{}))
				Expect(reports.InstanceReports[0].LastSpike.To).To(Equal(time.Time{}))
				Expect(reports.InstanceReports[1].LastSpike.From).To(Equal(time.Unix(1, 0)))
				Expect(reports.InstanceReports[1].LastSpike.To).To(Equal(time.Unix(2, 0)))
			})
		})

		When("some instances have spiked", func() {
			BeforeEach(func() {
				lastSpikeFetcher.FetchInstanceDataReturns(map[int]interface{}{
					0: fetchers.LastSpikeInstanceData{
						InstanceID: 0,
						From:       time.Unix(3, 0),
						To:         time.Unix(5, 0),
					},
				}, nil)
			})

			It("adds the spike starting and ending times to the report", func() {
				Expect(reports.InstanceReports).To(HaveLen(1))
				Expect(reports.InstanceReports[0].LastSpike.From).To(Equal(time.Unix(3, 0)))
				Expect(reports.InstanceReports[0].LastSpike.To).To(Equal(time.Unix(5, 0)))
			})
		})
	})

	Describe("Current CPU usage", func() {
		BeforeEach(func() {
			currentUsageFetcher.FetchInstanceDataReturns(map[int]interface{}{
				0: fetchers.CurrentInstanceData{
					InstanceID: 0,
					Usage:      1.5,
				},
				1: fetchers.CurrentInstanceData{
					InstanceID: 1,
					Usage:      1.7,
				},
			}, nil)
		})

		It("fetches the usage data correctly", func() {
			Expect(currentUsageFetcher.FetchInstanceDataCallCount()).To(Equal(1))
			actualAppGuid, actualAppInstances := currentUsageFetcher.FetchInstanceDataArgsForCall(0)
			Expect(actualAppGuid).To(Equal(appGuid))
			Expect(actualAppInstances).To(Equal(appInstances))
		})

		When("fetching the current usage fails", func() {
			BeforeEach(func() {
				currentUsageFetcher.FetchInstanceDataReturns(nil, errors.New("fetch-current-error"))
			})

			It("returns the error", func() {
				Expect(err).To(MatchError("fetch-current-error"))
			})
		})

		When("the fetcher returns the wrong type of instance data", func() {
			BeforeEach(func() {
				currentUsageFetcher.FetchInstanceDataReturns(map[int]interface{}{
					0: "hello",
					1: fetchers.CurrentInstanceData{Usage: 1.7},
				}, nil)
			})

			It("skips the instance with the wrong type", func() {
				Expect(len(reports.InstanceReports)).To(Equal(1))
				Expect(reports.InstanceReports[0].CurrentUsage).To(Equal(reporter.CurrentUsage{Value: 1.7}))
			})
		})

		It("reports current usage", func() {
			Expect(len(reports.InstanceReports)).To(Equal(2))

			Expect(reports.InstanceReports[0].InstanceID).To(Equal(0))
			Expect(reports.InstanceReports[0].CurrentUsage.Value).To(Equal(1.5))

			Expect(reports.InstanceReports[1].InstanceID).To(Equal(1))
			Expect(reports.InstanceReports[1].CurrentUsage.Value).To(Equal(1.7))
		})
	})
})
