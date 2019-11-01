package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"code.cloudfoundry.org/cpu-entitlement-plugin/cf"
	"code.cloudfoundry.org/cpu-entitlement-plugin/fetchers"
	"code.cloudfoundry.org/cpu-entitlement-plugin/httpclient"
	logcache "code.cloudfoundry.org/log-cache/pkg/client"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = FDescribe("Last Spike Fetcher", func() {
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

	It("returns the last spike", func() {
		fmt.Printf("appGuid = %+v\n", appGuid)

		firstSpike := map[string]string{
			"source_id":           appGuid,
			"instance_id":         "0",
			"process_instance_id": "1",
			"spike_start":         "2019-11-01T00:02:00Z",
			"spike_end":           "2019-11-01T00:03:00Z",
		}

		secondSpike := map[string]string{
			"source_id":           appGuid,
			"instance_id":         "0",
			"process_instance_id": "1",
			"spike_start":         "2019-11-01T10:02:00Z",
			"spike_end":           "2019-11-01T10:03:00Z",
		}

		emitSpike(firstSpike)
		emitSpike(secondSpike)

		spikes, err := fetcher.FetchInstanceData(appGuid, map[int]cf.Instance{0: cf.Instance{InstanceID: 0, ProcessInstanceID: "1"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(len(spikes)).To(Equal(1))

		expectedFrom, err := time.Parse(time.RFC3339, secondSpike["spike_start"])
		Expect(err).NotTo(HaveOccurred())
		expectedTo, err := time.Parse(time.RFC3339, secondSpike["spike_end"])
		Expect(err).NotTo(HaveOccurred())

		Expect(spikes).To(HaveKey(0))

		spike := spikes[0].(fetchers.LastSpikeInstanceData)
		Expect(spike.InstanceID).To(Equal(0))
		Expect(spike.From).To(BeTemporally("==", expectedFrom))
		Expect(spike.To).To(BeTemporally("==", expectedTo))
	})
})

func emitSpike(spike map[string]string) {
	spikeBytes, err := json.Marshal(spike)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	response, err := logEmitterHttpClient.Post(getTestLogEmitterURL()+"/spike", "application/json", bytes.NewBuffer(spikeBytes))
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	ExpectWithOffset(1, response.StatusCode).To(Equal(http.StatusOK), string(responseBody))
}
