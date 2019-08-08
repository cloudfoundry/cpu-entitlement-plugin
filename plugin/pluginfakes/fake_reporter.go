// Code generated by counterfeiter. DO NOT EDIT.
package pluginfakes

import (
	"sync"

	"code.cloudfoundry.org/cpu-entitlement-plugin/metadata"
	"code.cloudfoundry.org/cpu-entitlement-plugin/plugin"
	"code.cloudfoundry.org/cpu-entitlement-plugin/reporter"
)

type FakeReporter struct {
	CreateInstanceReportsStub        func(metadata.CFAppInfo) ([]reporter.InstanceReport, error)
	createInstanceReportsMutex       sync.RWMutex
	createInstanceReportsArgsForCall []struct {
		arg1 metadata.CFAppInfo
	}
	createInstanceReportsReturns struct {
		result1 []reporter.InstanceReport
		result2 error
	}
	createInstanceReportsReturnsOnCall map[int]struct {
		result1 []reporter.InstanceReport
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeReporter) CreateInstanceReports(arg1 metadata.CFAppInfo) ([]reporter.InstanceReport, error) {
	fake.createInstanceReportsMutex.Lock()
	ret, specificReturn := fake.createInstanceReportsReturnsOnCall[len(fake.createInstanceReportsArgsForCall)]
	fake.createInstanceReportsArgsForCall = append(fake.createInstanceReportsArgsForCall, struct {
		arg1 metadata.CFAppInfo
	}{arg1})
	fake.recordInvocation("CreateInstanceReports", []interface{}{arg1})
	fake.createInstanceReportsMutex.Unlock()
	if fake.CreateInstanceReportsStub != nil {
		return fake.CreateInstanceReportsStub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.createInstanceReportsReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeReporter) CreateInstanceReportsCallCount() int {
	fake.createInstanceReportsMutex.RLock()
	defer fake.createInstanceReportsMutex.RUnlock()
	return len(fake.createInstanceReportsArgsForCall)
}

func (fake *FakeReporter) CreateInstanceReportsCalls(stub func(metadata.CFAppInfo) ([]reporter.InstanceReport, error)) {
	fake.createInstanceReportsMutex.Lock()
	defer fake.createInstanceReportsMutex.Unlock()
	fake.CreateInstanceReportsStub = stub
}

func (fake *FakeReporter) CreateInstanceReportsArgsForCall(i int) metadata.CFAppInfo {
	fake.createInstanceReportsMutex.RLock()
	defer fake.createInstanceReportsMutex.RUnlock()
	argsForCall := fake.createInstanceReportsArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeReporter) CreateInstanceReportsReturns(result1 []reporter.InstanceReport, result2 error) {
	fake.createInstanceReportsMutex.Lock()
	defer fake.createInstanceReportsMutex.Unlock()
	fake.CreateInstanceReportsStub = nil
	fake.createInstanceReportsReturns = struct {
		result1 []reporter.InstanceReport
		result2 error
	}{result1, result2}
}

func (fake *FakeReporter) CreateInstanceReportsReturnsOnCall(i int, result1 []reporter.InstanceReport, result2 error) {
	fake.createInstanceReportsMutex.Lock()
	defer fake.createInstanceReportsMutex.Unlock()
	fake.CreateInstanceReportsStub = nil
	if fake.createInstanceReportsReturnsOnCall == nil {
		fake.createInstanceReportsReturnsOnCall = make(map[int]struct {
			result1 []reporter.InstanceReport
			result2 error
		})
	}
	fake.createInstanceReportsReturnsOnCall[i] = struct {
		result1 []reporter.InstanceReport
		result2 error
	}{result1, result2}
}

func (fake *FakeReporter) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.createInstanceReportsMutex.RLock()
	defer fake.createInstanceReportsMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeReporter) recordInvocation(key string, args []interface{}) {
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

var _ plugin.Reporter = new(FakeReporter)
