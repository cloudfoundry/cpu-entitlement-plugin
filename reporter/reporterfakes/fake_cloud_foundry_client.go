// Code generated by counterfeiter. DO NOT EDIT.
package reporterfakes

import (
	"sync"

	"code.cloudfoundry.org/cpu-entitlement-plugin/cf"
	"code.cloudfoundry.org/cpu-entitlement-plugin/reporter"
)

type FakeCloudFoundryClient struct {
	GetCurrentOrgStub        func() (string, error)
	getCurrentOrgMutex       sync.RWMutex
	getCurrentOrgArgsForCall []struct {
	}
	getCurrentOrgReturns struct {
		result1 string
		result2 error
	}
	getCurrentOrgReturnsOnCall map[int]struct {
		result1 string
		result2 error
	}
	GetSpacesStub        func() ([]cf.Space, error)
	getSpacesMutex       sync.RWMutex
	getSpacesArgsForCall []struct {
	}
	getSpacesReturns struct {
		result1 []cf.Space
		result2 error
	}
	getSpacesReturnsOnCall map[int]struct {
		result1 []cf.Space
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeCloudFoundryClient) GetCurrentOrg() (string, error) {
	fake.getCurrentOrgMutex.Lock()
	ret, specificReturn := fake.getCurrentOrgReturnsOnCall[len(fake.getCurrentOrgArgsForCall)]
	fake.getCurrentOrgArgsForCall = append(fake.getCurrentOrgArgsForCall, struct {
	}{})
	fake.recordInvocation("GetCurrentOrg", []interface{}{})
	fake.getCurrentOrgMutex.Unlock()
	if fake.GetCurrentOrgStub != nil {
		return fake.GetCurrentOrgStub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.getCurrentOrgReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeCloudFoundryClient) GetCurrentOrgCallCount() int {
	fake.getCurrentOrgMutex.RLock()
	defer fake.getCurrentOrgMutex.RUnlock()
	return len(fake.getCurrentOrgArgsForCall)
}

func (fake *FakeCloudFoundryClient) GetCurrentOrgCalls(stub func() (string, error)) {
	fake.getCurrentOrgMutex.Lock()
	defer fake.getCurrentOrgMutex.Unlock()
	fake.GetCurrentOrgStub = stub
}

func (fake *FakeCloudFoundryClient) GetCurrentOrgReturns(result1 string, result2 error) {
	fake.getCurrentOrgMutex.Lock()
	defer fake.getCurrentOrgMutex.Unlock()
	fake.GetCurrentOrgStub = nil
	fake.getCurrentOrgReturns = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *FakeCloudFoundryClient) GetCurrentOrgReturnsOnCall(i int, result1 string, result2 error) {
	fake.getCurrentOrgMutex.Lock()
	defer fake.getCurrentOrgMutex.Unlock()
	fake.GetCurrentOrgStub = nil
	if fake.getCurrentOrgReturnsOnCall == nil {
		fake.getCurrentOrgReturnsOnCall = make(map[int]struct {
			result1 string
			result2 error
		})
	}
	fake.getCurrentOrgReturnsOnCall[i] = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *FakeCloudFoundryClient) GetSpaces() ([]cf.Space, error) {
	fake.getSpacesMutex.Lock()
	ret, specificReturn := fake.getSpacesReturnsOnCall[len(fake.getSpacesArgsForCall)]
	fake.getSpacesArgsForCall = append(fake.getSpacesArgsForCall, struct {
	}{})
	fake.recordInvocation("GetSpaces", []interface{}{})
	fake.getSpacesMutex.Unlock()
	if fake.GetSpacesStub != nil {
		return fake.GetSpacesStub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.getSpacesReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeCloudFoundryClient) GetSpacesCallCount() int {
	fake.getSpacesMutex.RLock()
	defer fake.getSpacesMutex.RUnlock()
	return len(fake.getSpacesArgsForCall)
}

func (fake *FakeCloudFoundryClient) GetSpacesCalls(stub func() ([]cf.Space, error)) {
	fake.getSpacesMutex.Lock()
	defer fake.getSpacesMutex.Unlock()
	fake.GetSpacesStub = stub
}

func (fake *FakeCloudFoundryClient) GetSpacesReturns(result1 []cf.Space, result2 error) {
	fake.getSpacesMutex.Lock()
	defer fake.getSpacesMutex.Unlock()
	fake.GetSpacesStub = nil
	fake.getSpacesReturns = struct {
		result1 []cf.Space
		result2 error
	}{result1, result2}
}

func (fake *FakeCloudFoundryClient) GetSpacesReturnsOnCall(i int, result1 []cf.Space, result2 error) {
	fake.getSpacesMutex.Lock()
	defer fake.getSpacesMutex.Unlock()
	fake.GetSpacesStub = nil
	if fake.getSpacesReturnsOnCall == nil {
		fake.getSpacesReturnsOnCall = make(map[int]struct {
			result1 []cf.Space
			result2 error
		})
	}
	fake.getSpacesReturnsOnCall[i] = struct {
		result1 []cf.Space
		result2 error
	}{result1, result2}
}

func (fake *FakeCloudFoundryClient) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.getCurrentOrgMutex.RLock()
	defer fake.getCurrentOrgMutex.RUnlock()
	fake.getSpacesMutex.RLock()
	defer fake.getSpacesMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeCloudFoundryClient) recordInvocation(key string, args []interface{}) {
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

var _ reporter.CloudFoundryClient = new(FakeCloudFoundryClient)