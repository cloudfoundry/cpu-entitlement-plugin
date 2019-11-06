package emitters

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"code.cloudfoundry.org/go-loggregator"
)

type CounterMetric struct {
	Name       string
	SourceId   string
	InstanceId string
	Tags       map[string]string
}

type CounterEmitter struct {
	client *loggregator.IngressClient
}

func NewCounterEmitter(client *loggregator.IngressClient) *CounterEmitter {
	return &CounterEmitter{client: client}
}

func (e CounterEmitter) SendCounter(counter CounterMetric) {
	e.client.EmitCounter(counter.Name,
		loggregator.WithCounterSourceInfo(counter.SourceId, counter.InstanceId),
		loggregator.WithEnvelopeTags(counter.Tags),
	)
}

func (e CounterEmitter) EmitCounter() http.HandlerFunc {
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

		counterMetric := CounterMetric{}
		if err := json.Unmarshal(body, &counterMetric); err != nil {
			http.Error(w, fmt.Sprintf("Failed to unmarshal body: %v", err), http.StatusInternalServerError)
			return
		}

		e.SendCounter(counterMetric)
	}
}
