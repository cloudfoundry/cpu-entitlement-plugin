package fetchers_test

import (
	"testing"

	"code.cloudfoundry.org/log-cache/pkg/rpc/logcache_v1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFetchers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fetchers Suite")
}

type Metric struct {
	Usage       float64
	Entitlement float64
	Age         float64
}

func rangeQueryResult(series ...*logcache_v1.PromQL_Series) *logcache_v1.PromQL_RangeQueryResult {
	return &logcache_v1.PromQL_RangeQueryResult{
		Result: &logcache_v1.PromQL_RangeQueryResult_Matrix{
			Matrix: &logcache_v1.PromQL_Matrix{
				Series: series,
			},
		},
	}
}

func queryResult(samples ...*logcache_v1.PromQL_Sample) *logcache_v1.PromQL_InstantQueryResult {
	return &logcache_v1.PromQL_InstantQueryResult{
		Result: &logcache_v1.PromQL_InstantQueryResult_Vector{
			Vector: &logcache_v1.PromQL_Vector{
				Samples: samples,
			},
		},
	}
}

func series(instanceID, procInstanceID string, points ...*logcache_v1.PromQL_Point) *logcache_v1.PromQL_Series {
	return &logcache_v1.PromQL_Series{
		Metric: map[string]string{
			"instance_id":         instanceID,
			"process_instance_id": procInstanceID,
		},
		Points: points,
	}
}

func sample(instanceID, procInstanceID string, point *logcache_v1.PromQL_Point) *logcache_v1.PromQL_Sample {
	return &logcache_v1.PromQL_Sample{
		Metric: map[string]string{
			"instance_id":         instanceID,
			"process_instance_id": procInstanceID,
		},
		Point: point,
	}
}

func point(time string, value float64) *logcache_v1.PromQL_Point {
	return &logcache_v1.PromQL_Point{Time: time, Value: value}
}
