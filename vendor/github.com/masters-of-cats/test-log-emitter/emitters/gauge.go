package emitters

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"code.cloudfoundry.org/go-loggregator"
)

type GaugeValue struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

type GaugeMetric struct {
	SourceId   string
	InstanceId string
	Tags       map[string]string
	Values     []GaugeValue
}

type GaugeEmitter struct {
	client *loggregator.IngressClient
}

func NewGaugeEmitter(client *loggregator.IngressClient) *GaugeEmitter {
	return &GaugeEmitter{client: client}
}

func (e GaugeEmitter) SendGauge(gauge GaugeMetric) {
	opts := []loggregator.EmitGaugeOption{
		loggregator.WithGaugeSourceInfo(gauge.SourceId, gauge.InstanceId),
		loggregator.WithEnvelopeTags(gauge.Tags),
	}
	for _, val := range gauge.Values {
		opts = append(opts, loggregator.WithGaugeValue(val.Name, val.Value, val.Unit))
	}
	e.client.EmitGauge(opts...)
}

func (e GaugeEmitter) EmitGauge() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Sorry, only POST methods are supported.", http.StatusMethodNotAllowed)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read body: %v", err), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		gaugeMetric := GaugeMetric{}
		if err := json.Unmarshal(body, &gaugeMetric); err != nil {
			http.Error(w, fmt.Sprintf("Failed to unmarshal body: %v", err), http.StatusInternalServerError)
			return
		}

		e.SendGauge(gaugeMetric)
	}
}
