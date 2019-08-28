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
		historicalUsageFetcher *reporterfakes.FakeInstanceDataFetcher
		currentUsageFetcher    *reporterfakes.FakeInstanceDataFetcher
		instanceReporter       reporter.AppReporter
		reports                []reporter.InstanceReport
		err                    error
		appGuid                string
	)

	BeforeEach(func() {
		appGuid = "foo"

		historicalUsageFetcher = new(reporterfakes.FakeInstanceDataFetcher)
		currentUsageFetcher = new(reporterfakes.FakeInstanceDataFetcher)
		instanceReporter = reporter.NewAppReporter(historicalUsageFetcher, currentUsageFetcher)
	})

	JustBeforeEach(func() {
		reports, err = instanceReporter.CreateInstanceReports(cf.Application{Guid: appGuid})
	})

	Describe("Report", func() {
		BeforeEach(func() {
			historicalUsageFetcher.FetchInstanceDataReturns(map[int][]fetchers.InstanceData{
				0: {
					{
						InstanceID: 0,
						Value:      0.5,
					},
				},
			}, nil)
			currentUsageFetcher.FetchInstanceDataReturns(map[int][]fetchers.InstanceData{
				0: {
					{
						InstanceID: 0,
						Value:      1.5,
					},
				},
			}, nil)
		})

		It("combines historical and current cpu usage data", func() {
			Expect(len(reports)).To(Equal(1))

			Expect(reports[0].InstanceID).To(Equal(0))
			Expect(reports[0].HistoricalUsage.Value).To(Equal(0.5))
			Expect(reports[0].CurrentUsage.Value).To(Equal(1.5))
		})

		When("current usage data and historical usage data cannot be matched by instance id", func() {
			BeforeEach(func() {
				historicalUsageFetcher.FetchInstanceDataReturns(map[int][]fetchers.InstanceData{
					0: {
						{
							InstanceID: 0,
							Value:      0.5,
						},
					},
				}, nil)
				currentUsageFetcher.FetchInstanceDataReturns(map[int][]fetchers.InstanceData{
					1: {
						{
							InstanceID: 1,
							Value:      1.5,
						},
					},
				}, nil)
			})

			It("combines historical and current cpu usage data", func() {
				Expect(len(reports)).To(Equal(2))

				Expect(reports[0].InstanceID).To(Equal(0))
				Expect(reports[0].HistoricalUsage.Value).To(Equal(0.5))
				Expect(reports[0].CurrentUsage.Value).To(BeZero())

				Expect(reports[1].InstanceID).To(Equal(1))
				Expect(reports[1].HistoricalUsage.Value).To(BeZero())
				Expect(reports[1].CurrentUsage.Value).To(Equal(1.5))
			})
		})
	})

	Describe("Historical CPU usage", func() {
		BeforeEach(func() {
			historicalUsageFetcher.FetchInstanceDataReturns(map[int][]fetchers.InstanceData{
				0: {
					{
						InstanceID: 0,
						Value:      0.5,
					},
				},
				1: {
					{
						InstanceID: 1,
						Value:      0.6,
					},
					{
						InstanceID: 1,
						Value:      0.7,
					},
				},
			}, nil)
		})
		It("fetches the usage data correctly", func() {
			Expect(historicalUsageFetcher.FetchInstanceDataCallCount()).To(Equal(1))
			Expect(historicalUsageFetcher.FetchInstanceDataArgsForCall(0)).To(Equal(appGuid))
		})

		When("fetching the historical usage fails", func() {
			BeforeEach(func() {
				historicalUsageFetcher.FetchInstanceDataReturns(nil, errors.New("fetch-historical-error"))
			})

			It("returns the error", func() {
				Expect(err).To(MatchError("fetch-historical-error"))
			})
		})

		It("calculates historical entitlement ratio", func() {
			Expect(len(reports)).To(Equal(2))

			Expect(reports[0].InstanceID).To(Equal(0))
			Expect(reports[0].HistoricalUsage.Value).To(Equal(0.5))

			Expect(reports[1].InstanceID).To(Equal(1))
			Expect(reports[1].HistoricalUsage.Value).To(Equal(0.7))
		})

		When("an instance is missing from the historical usage data", func() {
			BeforeEach(func() {
				historicalUsageFetcher.FetchInstanceDataReturns(map[int][]fetchers.InstanceData{
					2: {
						{
							InstanceID: 2,
							Value:      0.5,
						},
					},
					0: {
						{
							InstanceID: 0,
							Value:      0.6,
						},
					},
				}, nil)
				currentUsageFetcher.FetchInstanceDataReturns(map[int][]fetchers.InstanceData{}, nil)
			})

			It("still returns an (incomplete) result", func() {
				Expect(len(reports)).To(Equal(2))

				Expect(reports[0].InstanceID).To(Equal(0))
				Expect(reports[0].HistoricalUsage.Value).To(Equal(0.6))

				Expect(reports[1].InstanceID).To(Equal(2))
				Expect(reports[1].HistoricalUsage.Value).To(Equal(0.5))
			})
		})
		When("some instances have spiked", func() {
			BeforeEach(func() {
				historicalUsageFetcher.FetchInstanceDataReturns(map[int][]fetchers.InstanceData{
					0: {
						{InstanceID: 0, Time: time.Unix(1, 0), Value: 0.5},
						{InstanceID: 0, Time: time.Unix(3, 0), Value: 1.5},
						{InstanceID: 0, Time: time.Unix(5, 0), Value: 2.0},
						{InstanceID: 0, Time: time.Unix(6, 0), Value: 0.9},
					},
					1: {
						{InstanceID: 1, Time: time.Unix(2, 0), Value: 0.6},
						{InstanceID: 1, Time: time.Unix(4, 0), Value: 0.4},
					},
				}, nil)
			})

			It("adds the spike starting and ending times to the report", func() {
				Expect(reports[0].HistoricalUsage.LastSpikeFrom).To(Equal(time.Unix(3, 0)))
				Expect(reports[0].HistoricalUsage.LastSpikeTo).To(Equal(time.Unix(5, 0)))
			})
		})

		When("latest spike starts at beginning of data and ends before end of data", func() {
			BeforeEach(func() {
				historicalUsageFetcher.FetchInstanceDataReturns(map[int][]fetchers.InstanceData{
					0: {
						{InstanceID: 0, Time: time.Unix(1, 0), Value: 2.5},
						{InstanceID: 0, Time: time.Unix(2, 0), Value: 1.5},
						{InstanceID: 0, Time: time.Unix(3, 0), Value: 0.9},
					},
				}, nil)
			})

			It("reports spike from beginning of data to end of spike", func() {
				Expect(reports[0].HistoricalUsage.LastSpikeFrom).To(Equal(time.Unix(1, 0)))
				Expect(reports[0].HistoricalUsage.LastSpikeTo).To(Equal(time.Unix(2, 0)))
			})
		})

		When("latest spike starts at beginning of data and is always spiking in range", func() {
			BeforeEach(func() {
				historicalUsageFetcher.FetchInstanceDataReturns(map[int][]fetchers.InstanceData{
					0: {
						{InstanceID: 0, Time: time.Unix(1, 0), Value: 1.5},
						{InstanceID: 0, Time: time.Unix(2, 0), Value: 2.5},
					},
				}, nil)
			})

			It("reports spike from beginning of data to end of data", func() {
				Expect(reports[0].HistoricalUsage.LastSpikeFrom).To(Equal(time.Unix(1, 0)))
				Expect(reports[0].HistoricalUsage.LastSpikeTo).To(Equal(time.Unix(2, 0)))
			})
		})

		When("latest spike is spiking at end of data", func() {
			BeforeEach(func() {
				historicalUsageFetcher.FetchInstanceDataReturns(map[int][]fetchers.InstanceData{
					0: {
						{InstanceID: 0, Time: time.Unix(1, 0), Value: 0.5},
						{InstanceID: 0, Time: time.Unix(2, 0), Value: 1.5},
						{InstanceID: 0, Time: time.Unix(3, 0), Value: 2.5},
					},
				}, nil)
			})

			It("reports spike from beginning of spike to end of data", func() {
				Expect(reports[0].HistoricalUsage.LastSpikeFrom).To(Equal(time.Unix(2, 0)))
				Expect(reports[0].HistoricalUsage.LastSpikeTo).To(Equal(time.Unix(3, 0)))
			})
		})

		When("multiple spikes exist", func() {
			BeforeEach(func() {
				historicalUsageFetcher.FetchInstanceDataReturns(map[int][]fetchers.InstanceData{
					0: {
						{InstanceID: 0, Time: time.Unix(2, 0), Value: 0.5},
						{InstanceID: 0, Time: time.Unix(3, 0), Value: 0.7},
						{InstanceID: 0, Time: time.Unix(4, 0), Value: 0.9},
						{InstanceID: 0, Time: time.Unix(5, 0), Value: 0.8},
						{InstanceID: 0, Time: time.Unix(6, 0), Value: 1.2},
						{InstanceID: 0, Time: time.Unix(7, 0), Value: 1.5},
					},
				}, nil)
			})

			It("reports only the latest spike", func() {
				Expect(reports[0].HistoricalUsage.LastSpikeFrom).To(Equal(time.Unix(6, 0)))
				Expect(reports[0].HistoricalUsage.LastSpikeTo).To(Equal(time.Unix(7, 0)))
			})
		})

		When("a spike consists of a single data point", func() {
			BeforeEach(func() {
				historicalUsageFetcher.FetchInstanceDataReturns(map[int][]fetchers.InstanceData{
					0: {
						{InstanceID: 0, Time: time.Unix(2, 0), Value: 0.8},
						{InstanceID: 0, Time: time.Unix(3, 0), Value: 1.5},
						{InstanceID: 0, Time: time.Unix(4, 0), Value: 0.5},
					},
				}, nil)
			})

			It("reports an empty range", func() {
				Expect(reports[0].HistoricalUsage.LastSpikeFrom).To(Equal(time.Unix(3, 0)))
				Expect(reports[0].HistoricalUsage.LastSpikeTo).To(Equal(time.Unix(3, 0)))
			})
		})

		When("an instance reaches 100% entitlement usage but doesn't go above", func() {
			BeforeEach(func() {
				historicalUsageFetcher.FetchInstanceDataReturns(map[int][]fetchers.InstanceData{
					0: {
						{InstanceID: 0, Time: time.Unix(2, 0), Value: 0.5},
						{InstanceID: 0, Time: time.Unix(3, 0), Value: 1.0},
						{InstanceID: 0, Time: time.Unix(4, 0), Value: 0.8},
					},
				}, nil)
			})

			It("does not report a spike", func() {
				Expect(reports[0].HistoricalUsage.LastSpikeFrom.IsZero()).To(BeTrue())
				Expect(reports[0].HistoricalUsage.LastSpikeTo.IsZero()).To(BeTrue())
			})
		})
	})

	Describe("Current CPU usage", func() {
		BeforeEach(func() {
			currentUsageFetcher.FetchInstanceDataReturns(map[int][]fetchers.InstanceData{
				0: {
					{
						InstanceID: 0,
						Value:      1.5,
					},
				},
				1: {
					{
						InstanceID: 1,
						Value:      1.7,
					},
				},
			}, nil)
		})

		It("fetches the usage data correctly", func() {
			Expect(currentUsageFetcher.FetchInstanceDataCallCount()).To(Equal(1))
			Expect(currentUsageFetcher.FetchInstanceDataArgsForCall(0)).To(Equal(appGuid))
		})

		When("fetching the current usage fails", func() {
			BeforeEach(func() {
				currentUsageFetcher.FetchInstanceDataReturns(nil, errors.New("fetch-current-error"))
			})

			It("returns the error", func() {
				Expect(err).To(MatchError("fetch-current-error"))
			})
		})

		It("calculates current entitlement ratio", func() {
			Expect(len(reports)).To(Equal(2))

			Expect(reports[0].InstanceID).To(Equal(0))
			Expect(reports[0].CurrentUsage.Value).To(Equal(1.5))

			Expect(reports[1].InstanceID).To(Equal(1))
			Expect(reports[1].CurrentUsage.Value).To(Equal(1.7))
		})
	})
})
