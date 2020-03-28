package integration_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"code.cloudfoundry.org/cpu-entitlement-plugin/cf"
	"code.cloudfoundry.org/cpu-entitlement-plugin/fetchers"
	"github.com/google/uuid"
	"github.com/masters-of-cats/test-log-emitter/emitters"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Last Spike Fetcher", func() {
	var (
		appGuid string
		fetcher *fetchers.LastSpikeFetcher
	)

	getSpikes := func(appGuid string, instanceMap map[int]cf.Instance) map[int]interface{} {
		spikes, err := fetcher.FetchInstanceData(logger, appGuid, instanceMap)
		Expect(err).NotTo(HaveOccurred())
		return spikes
	}

	BeforeEach(func() {
		uid := uuid.New().String()
		appGuid = "test-app-" + uid

		fetcher = fetchers.NewLastSpikeFetcher(logCacheClient, time.Now().Add(-30*24*time.Hour))
	})

	When("multiple spikes have been emitted", func() {
		BeforeEach(func() {
			emitGauge(emitters.GaugeMetric{
				SourceId:   appGuid,
				InstanceId: "0",
				Tags: map[string]string{
					"process_instance_id": "1",
				},
				Values: []emitters.GaugeValue{
					{Name: "spike_start", Value: 134, Unit: "seconds"},
					{Name: "spike_end", Value: 136, Unit: "seconds"},
				},
			})

			emitGauge(emitters.GaugeMetric{
				SourceId:   appGuid,
				InstanceId: "0",
				Tags: map[string]string{
					"process_instance_id": "1",
				},
				Values: []emitters.GaugeValue{
					{Name: "spike_start", Value: 234, Unit: "seconds"},
					{Name: "spike_end", Value: 236, Unit: "seconds"},
				},
			})
		})

		It("returns the most recent spike", func() {
			var spikes map[int]interface{}
			Eventually(func() map[int]interface{} {
				spikes = getSpikes(appGuid, map[int]cf.Instance{0: cf.Instance{InstanceID: 0, ProcessInstanceID: "1"}})
				return spikes
			}).Should(HaveLen(1))

			expectedFrom := time.Unix(234, 0)
			expectedTo := time.Unix(236, 0)

			Expect(spikes).To(HaveKey(0))

			spike := spikes[0].(fetchers.LastSpikeInstanceData)
			Expect(spike.InstanceID).To(Equal(0))
			Expect(spike.From).To(BeTemporally("==", expectedFrom))
			Expect(spike.To).To(BeTemporally("==", expectedTo))
		})
	})

	When("a non-spike gauge has been emitted", func() {
		BeforeEach(func() {
			emitGauge(emitters.GaugeMetric{
				SourceId:   appGuid,
				InstanceId: "0",
				Tags: map[string]string{
					"process_instance_id": "1",
				},
				Values: []emitters.GaugeValue{
					{Name: "spoke_start", Value: 134, Unit: "seconds"},
					{Name: "spoke_end", Value: 136, Unit: "seconds"},
				},
			})
		})

		It("ignores it", func() {
			spikes := getSpikes(appGuid, map[int]cf.Instance{0: cf.Instance{InstanceID: 0, ProcessInstanceID: "1"}})
			Expect(spikes).To(BeEmpty())
		})
	})

	When("a non-gauge metric called 'spike*' has been emitted", func() {
		BeforeEach(func() {
			emitCounter(emitters.CounterMetric{
				Name:       "spike_counter",
				SourceId:   appGuid,
				InstanceId: "0",
				Tags: map[string]string{
					"process_instance_id": "1",
				},
			})
		})

		It("ignores it", func() {
			spikes := getSpikes(appGuid, map[int]cf.Instance{0: cf.Instance{InstanceID: 0, ProcessInstanceID: "1"}})
			Expect(spikes).To(BeEmpty())
		})
	})

	When("a metric belongs to an old instance", func() {
		BeforeEach(func() {
			emitGauge(emitters.GaugeMetric{
				SourceId:   appGuid,
				InstanceId: "0",
				Tags: map[string]string{
					"process_instance_id": "old",
				},
				Values: []emitters.GaugeValue{
					{Name: "spike_start", Value: 234, Unit: "seconds"},
					{Name: "spike_end", Value: 236, Unit: "seconds"},
				},
			})
		})

		It("ignores it", func() {
			spikes := getSpikes(appGuid, map[int]cf.Instance{0: cf.Instance{InstanceID: 0, ProcessInstanceID: "1"}})
			Expect(spikes).To(BeEmpty())
		})
	})
})

func emitGauge(gauge emitters.GaugeMetric) {
	emitMetric("/gauge", gauge)
}

func emitCounter(counter emitters.CounterMetric) {
	emitMetric("/counter", counter)
}

func emitMetric(endpoint string, metric interface{}) {
	metricBytes, err := json.Marshal(metric)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	response, err := logEmitterHttpClient.Post(getTestLogEmitterURL()+endpoint, "application/json", bytes.NewReader(metricBytes))
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	ExpectWithOffset(1, response.StatusCode).To(Equal(http.StatusOK), string(responseBody))
}
