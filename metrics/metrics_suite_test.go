package metrics_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMetricfetcher(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Metricfetcher Suite")
}
