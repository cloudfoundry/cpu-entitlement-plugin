// Code generated by counterfeiter. DO NOT EDIT.
package pluginfakes

import (
	"sync"

	"code.cloudfoundry.org/cpu-entitlement-plugin/calculator"
	"code.cloudfoundry.org/cpu-entitlement-plugin/metrics"
	"code.cloudfoundry.org/cpu-entitlement-plugin/plugin"
)

type FakeMetricsCalculator struct {
	CalculateInstanceReportsStub        func(map[int][]metrics.InstanceData) []calculator.InstanceReport
	calculateInstanceReportsMutex       sync.RWMutex
	calculateInstanceReportsArgsForCall []struct {
		arg1 map[int][]metrics.InstanceData
	}
	calculateInstanceReportsReturns struct {
		result1 []calculator.InstanceReport
	}
	calculateInstanceReportsReturnsOnCall map[int]struct {
		result1 []calculator.InstanceReport
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeMetricsCalculator) CalculateInstanceReports(arg1 map[int][]metrics.InstanceData) []calculator.InstanceReport {
	fake.calculateInstanceReportsMutex.Lock()
	ret, specificReturn := fake.calculateInstanceReportsReturnsOnCall[len(fake.calculateInstanceReportsArgsForCall)]
	fake.calculateInstanceReportsArgsForCall = append(fake.calculateInstanceReportsArgsForCall, struct {
		arg1 map[int][]metrics.InstanceData
	}{arg1})
	fake.recordInvocation("CalculateInstanceReports", []interface{}{arg1})
	fake.calculateInstanceReportsMutex.Unlock()
	if fake.CalculateInstanceReportsStub != nil {
		return fake.CalculateInstanceReportsStub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.calculateInstanceReportsReturns
	return fakeReturns.result1
}

func (fake *FakeMetricsCalculator) CalculateInstanceReportsCallCount() int {
	fake.calculateInstanceReportsMutex.RLock()
	defer fake.calculateInstanceReportsMutex.RUnlock()
	return len(fake.calculateInstanceReportsArgsForCall)
}

func (fake *FakeMetricsCalculator) CalculateInstanceReportsCalls(stub func(map[int][]metrics.InstanceData) []calculator.InstanceReport) {
	fake.calculateInstanceReportsMutex.Lock()
	defer fake.calculateInstanceReportsMutex.Unlock()
	fake.CalculateInstanceReportsStub = stub
}

func (fake *FakeMetricsCalculator) CalculateInstanceReportsArgsForCall(i int) map[int][]metrics.InstanceData {
	fake.calculateInstanceReportsMutex.RLock()
	defer fake.calculateInstanceReportsMutex.RUnlock()
	argsForCall := fake.calculateInstanceReportsArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeMetricsCalculator) CalculateInstanceReportsReturns(result1 []calculator.InstanceReport) {
	fake.calculateInstanceReportsMutex.Lock()
	defer fake.calculateInstanceReportsMutex.Unlock()
	fake.CalculateInstanceReportsStub = nil
	fake.calculateInstanceReportsReturns = struct {
		result1 []calculator.InstanceReport
	}{result1}
}

func (fake *FakeMetricsCalculator) CalculateInstanceReportsReturnsOnCall(i int, result1 []calculator.InstanceReport) {
	fake.calculateInstanceReportsMutex.Lock()
	defer fake.calculateInstanceReportsMutex.Unlock()
	fake.CalculateInstanceReportsStub = nil
	if fake.calculateInstanceReportsReturnsOnCall == nil {
		fake.calculateInstanceReportsReturnsOnCall = make(map[int]struct {
			result1 []calculator.InstanceReport
		})
	}
	fake.calculateInstanceReportsReturnsOnCall[i] = struct {
		result1 []calculator.InstanceReport
	}{result1}
}

func (fake *FakeMetricsCalculator) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.calculateInstanceReportsMutex.RLock()
	defer fake.calculateInstanceReportsMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeMetricsCalculator) recordInvocation(key string, args []interface{}) {
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

var _ plugin.MetricsCalculator = new(FakeMetricsCalculator)
