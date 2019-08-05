// Code generated by counterfeiter. DO NOT EDIT.
package outputfakes

import (
	"sync"

	"code.cloudfoundry.org/cpu-entitlement-plugin/output"
)

type FakeDisplay struct {
	ShowMessageStub        func(string, ...interface{})
	showMessageMutex       sync.RWMutex
	showMessageArgsForCall []struct {
		arg1 string
		arg2 []interface{}
	}
	ShowTableStub        func([]string, [][]string) error
	showTableMutex       sync.RWMutex
	showTableArgsForCall []struct {
		arg1 []string
		arg2 [][]string
	}
	showTableReturns struct {
		result1 error
	}
	showTableReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeDisplay) ShowMessage(arg1 string, arg2 ...interface{}) {
	fake.showMessageMutex.Lock()
	fake.showMessageArgsForCall = append(fake.showMessageArgsForCall, struct {
		arg1 string
		arg2 []interface{}
	}{arg1, arg2})
	fake.recordInvocation("ShowMessage", []interface{}{arg1, arg2})
	fake.showMessageMutex.Unlock()
	if fake.ShowMessageStub != nil {
		fake.ShowMessageStub(arg1, arg2...)
	}
}

func (fake *FakeDisplay) ShowMessageCallCount() int {
	fake.showMessageMutex.RLock()
	defer fake.showMessageMutex.RUnlock()
	return len(fake.showMessageArgsForCall)
}

func (fake *FakeDisplay) ShowMessageCalls(stub func(string, ...interface{})) {
	fake.showMessageMutex.Lock()
	defer fake.showMessageMutex.Unlock()
	fake.ShowMessageStub = stub
}

func (fake *FakeDisplay) ShowMessageArgsForCall(i int) (string, []interface{}) {
	fake.showMessageMutex.RLock()
	defer fake.showMessageMutex.RUnlock()
	argsForCall := fake.showMessageArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeDisplay) ShowTable(arg1 []string, arg2 [][]string) error {
	var arg1Copy []string
	if arg1 != nil {
		arg1Copy = make([]string, len(arg1))
		copy(arg1Copy, arg1)
	}
	var arg2Copy [][]string
	if arg2 != nil {
		arg2Copy = make([][]string, len(arg2))
		copy(arg2Copy, arg2)
	}
	fake.showTableMutex.Lock()
	ret, specificReturn := fake.showTableReturnsOnCall[len(fake.showTableArgsForCall)]
	fake.showTableArgsForCall = append(fake.showTableArgsForCall, struct {
		arg1 []string
		arg2 [][]string
	}{arg1Copy, arg2Copy})
	fake.recordInvocation("ShowTable", []interface{}{arg1Copy, arg2Copy})
	fake.showTableMutex.Unlock()
	if fake.ShowTableStub != nil {
		return fake.ShowTableStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.showTableReturns
	return fakeReturns.result1
}

func (fake *FakeDisplay) ShowTableCallCount() int {
	fake.showTableMutex.RLock()
	defer fake.showTableMutex.RUnlock()
	return len(fake.showTableArgsForCall)
}

func (fake *FakeDisplay) ShowTableCalls(stub func([]string, [][]string) error) {
	fake.showTableMutex.Lock()
	defer fake.showTableMutex.Unlock()
	fake.ShowTableStub = stub
}

func (fake *FakeDisplay) ShowTableArgsForCall(i int) ([]string, [][]string) {
	fake.showTableMutex.RLock()
	defer fake.showTableMutex.RUnlock()
	argsForCall := fake.showTableArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeDisplay) ShowTableReturns(result1 error) {
	fake.showTableMutex.Lock()
	defer fake.showTableMutex.Unlock()
	fake.ShowTableStub = nil
	fake.showTableReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeDisplay) ShowTableReturnsOnCall(i int, result1 error) {
	fake.showTableMutex.Lock()
	defer fake.showTableMutex.Unlock()
	fake.ShowTableStub = nil
	if fake.showTableReturnsOnCall == nil {
		fake.showTableReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.showTableReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeDisplay) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.showMessageMutex.RLock()
	defer fake.showMessageMutex.RUnlock()
	fake.showTableMutex.RLock()
	defer fake.showTableMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeDisplay) recordInvocation(key string, args []interface{}) {
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

var _ output.Display = new(FakeDisplay)