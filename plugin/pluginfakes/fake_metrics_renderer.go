// Code generated by counterfeiter. DO NOT EDIT.
package pluginfakes

import (
	"sync"

	"code.cloudfoundry.org/cpu-entitlement-plugin/metadata"
	"code.cloudfoundry.org/cpu-entitlement-plugin/metrics"
	"code.cloudfoundry.org/cpu-entitlement-plugin/plugin"
)

type FakeMetricsRenderer struct {
	ShowMetricsStub        func(metadata.CFAppInfo, []metrics.Usage) error
	showMetricsMutex       sync.RWMutex
	showMetricsArgsForCall []struct {
		arg1 metadata.CFAppInfo
		arg2 []metrics.Usage
	}
	showMetricsReturns struct {
		result1 error
	}
	showMetricsReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeMetricsRenderer) ShowMetrics(arg1 metadata.CFAppInfo, arg2 []metrics.Usage) error {
	var arg2Copy []metrics.Usage
	if arg2 != nil {
		arg2Copy = make([]metrics.Usage, len(arg2))
		copy(arg2Copy, arg2)
	}
	fake.showMetricsMutex.Lock()
	ret, specificReturn := fake.showMetricsReturnsOnCall[len(fake.showMetricsArgsForCall)]
	fake.showMetricsArgsForCall = append(fake.showMetricsArgsForCall, struct {
		arg1 metadata.CFAppInfo
		arg2 []metrics.Usage
	}{arg1, arg2Copy})
	fake.recordInvocation("ShowMetrics", []interface{}{arg1, arg2Copy})
	fake.showMetricsMutex.Unlock()
	if fake.ShowMetricsStub != nil {
		return fake.ShowMetricsStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.showMetricsReturns
	return fakeReturns.result1
}

func (fake *FakeMetricsRenderer) ShowMetricsCallCount() int {
	fake.showMetricsMutex.RLock()
	defer fake.showMetricsMutex.RUnlock()
	return len(fake.showMetricsArgsForCall)
}

func (fake *FakeMetricsRenderer) ShowMetricsCalls(stub func(metadata.CFAppInfo, []metrics.Usage) error) {
	fake.showMetricsMutex.Lock()
	defer fake.showMetricsMutex.Unlock()
	fake.ShowMetricsStub = stub
}

func (fake *FakeMetricsRenderer) ShowMetricsArgsForCall(i int) (metadata.CFAppInfo, []metrics.Usage) {
	fake.showMetricsMutex.RLock()
	defer fake.showMetricsMutex.RUnlock()
	argsForCall := fake.showMetricsArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeMetricsRenderer) ShowMetricsReturns(result1 error) {
	fake.showMetricsMutex.Lock()
	defer fake.showMetricsMutex.Unlock()
	fake.ShowMetricsStub = nil
	fake.showMetricsReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeMetricsRenderer) ShowMetricsReturnsOnCall(i int, result1 error) {
	fake.showMetricsMutex.Lock()
	defer fake.showMetricsMutex.Unlock()
	fake.ShowMetricsStub = nil
	if fake.showMetricsReturnsOnCall == nil {
		fake.showMetricsReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.showMetricsReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeMetricsRenderer) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.showMetricsMutex.RLock()
	defer fake.showMetricsMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeMetricsRenderer) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ plugin.MetricsRenderer = new(FakeMetricsRenderer)
