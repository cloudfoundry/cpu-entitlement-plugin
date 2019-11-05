package integration_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"code.cloudfoundry.org/cpu-entitlement-plugin/cf"
	"code.cloudfoundry.org/cpu-entitlement-plugin/fetchers"
	"code.cloudfoundry.org/cpu-entitlement-plugin/httpclient"
	logcache "code.cloudfoundry.org/log-cache/pkg/client"
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

	BeforeEach(func() {
		uid := uuid.New().String()
		appGuid = "test-app-" + uid

		logCacheURL := strings.Replace(cfApi, "https://api.", "http://log-cache.", 1)
		getToken := func() (string, error) {
			return getCmdOutput("cf", "oauth-token"), nil
		}

		logCacheClient := logcache.NewClient(
			logCacheURL,
			logcache.WithHTTPClient(httpclient.NewAuthClient(getToken)),
		)

		fetcher = fetchers.NewLastSpikeFetcher(logCacheClient, time.Now().Add(-30*24*time.Hour))
	})

	It("returns the most recent spike", func() {

		firstSpike := emitters.GaugeMetric{
			SourceId:   appGuid,
			InstanceId: "0",
			Tags: map[string]string{
				"process_instance_id": "1",
			},
			Values: []emitters.GaugeValue{
				{Name: "spike_start", Value: 134, Unit: "seconds"},
				{Name: "spike_end", Value: 136, Unit: "seconds"},
			},
		}

		secondSpike := emitters.GaugeMetric{
			SourceId:   appGuid,
			InstanceId: "0",
			Tags: map[string]string{
				"process_instance_id": "1",
			},
			Values: []emitters.GaugeValue{
				{Name: "spike_start", Value: 234, Unit: "seconds"},
				{Name: "spike_end", Value: 236, Unit: "seconds"},
			},
		}

		emitGauge(firstSpike)
		emitGauge(secondSpike)

		spikes, err := fetcher.FetchInstanceData(appGuid, map[int]cf.Instance{0: cf.Instance{InstanceID: 0, ProcessInstanceID: "1"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(len(spikes)).To(Equal(1))

		expectedFrom := time.Unix(234, 0)
		expectedTo := time.Unix(236, 0)

		Expect(spikes).To(HaveKey(0))

		spike := spikes[0].(fetchers.LastSpikeInstanceData)
		Expect(spike.InstanceID).To(Equal(0))
		Expect(spike.From).To(BeTemporally("==", expectedFrom))
		Expect(spike.To).To(BeTemporally("==", expectedTo))
	})

	It("doesn't return anything when we emit a spoke rather than a spike", func() {
		spoke := emitters.GaugeMetric{
			SourceId:   appGuid,
			InstanceId: "0",
			Tags: map[string]string{
				"process_instance_id": "1",
			},
			Values: []emitters.GaugeValue{
				{Name: "spoke_start", Value: 134, Unit: "seconds"},
				{Name: "spoke_end", Value: 136, Unit: "seconds"},
			},
		}

		emitGauge(spoke)
		spikes, err := fetcher.FetchInstanceData(appGuid, map[int]cf.Instance{0: cf.Instance{InstanceID: 0, ProcessInstanceID: "1"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(spikes).To(BeEmpty())
	})
})

func emitGauge(gauge emitters.GaugeMetric) {
	gaugeBytes, err := json.Marshal(gauge)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	response, err := logEmitterHttpClient.Post(getTestLogEmitterURL()+"/gauge", "application/json", bytes.NewReader(gaugeBytes))
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	ExpectWithOffset(1, response.StatusCode).To(Equal(http.StatusOK), string(responseBody))
}
