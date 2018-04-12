// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"encoding/json"
	"sync"

	"github.com/pivotal-cf/om/api"
)

type DirectorService struct {
	SetAZConfigurationStub        func(api.AZConfiguration) error
	setAZConfigurationMutex       sync.RWMutex
	setAZConfigurationArgsForCall []struct {
		arg1 api.AZConfiguration
	}
	setAZConfigurationReturns struct {
		result1 error
	}
	setAZConfigurationReturnsOnCall map[int]struct {
		result1 error
	}
	SetNetworksConfigurationStub        func(json.RawMessage) error
	setNetworksConfigurationMutex       sync.RWMutex
	setNetworksConfigurationArgsForCall []struct {
		arg1 json.RawMessage
	}
	setNetworksConfigurationReturns struct {
		result1 error
	}
	setNetworksConfigurationReturnsOnCall map[int]struct {
		result1 error
	}
	SetNetworkAndAZStub        func(api.NetworkAndAZConfiguration) error
	setNetworkAndAZMutex       sync.RWMutex
	setNetworkAndAZArgsForCall []struct {
		arg1 api.NetworkAndAZConfiguration
	}
	setNetworkAndAZReturns struct {
		result1 error
	}
	setNetworkAndAZReturnsOnCall map[int]struct {
		result1 error
	}
	SetPropertiesStub        func(api.DirectorProperties) error
	setPropertiesMutex       sync.RWMutex
	setPropertiesArgsForCall []struct {
		arg1 api.DirectorProperties
	}
	setPropertiesReturns struct {
		result1 error
	}
	setPropertiesReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *DirectorService) SetAZConfiguration(arg1 api.AZConfiguration) error {
	fake.setAZConfigurationMutex.Lock()
	ret, specificReturn := fake.setAZConfigurationReturnsOnCall[len(fake.setAZConfigurationArgsForCall)]
	fake.setAZConfigurationArgsForCall = append(fake.setAZConfigurationArgsForCall, struct {
		arg1 api.AZConfiguration
	}{arg1})
	fake.recordInvocation("SetAZConfiguration", []interface{}{arg1})
	fake.setAZConfigurationMutex.Unlock()
	if fake.SetAZConfigurationStub != nil {
		return fake.SetAZConfigurationStub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.setAZConfigurationReturns.result1
}

func (fake *DirectorService) SetAZConfigurationCallCount() int {
	fake.setAZConfigurationMutex.RLock()
	defer fake.setAZConfigurationMutex.RUnlock()
	return len(fake.setAZConfigurationArgsForCall)
}

func (fake *DirectorService) SetAZConfigurationArgsForCall(i int) api.AZConfiguration {
	fake.setAZConfigurationMutex.RLock()
	defer fake.setAZConfigurationMutex.RUnlock()
	return fake.setAZConfigurationArgsForCall[i].arg1
}

func (fake *DirectorService) SetAZConfigurationReturns(result1 error) {
	fake.SetAZConfigurationStub = nil
	fake.setAZConfigurationReturns = struct {
		result1 error
	}{result1}
}

func (fake *DirectorService) SetAZConfigurationReturnsOnCall(i int, result1 error) {
	fake.SetAZConfigurationStub = nil
	if fake.setAZConfigurationReturnsOnCall == nil {
		fake.setAZConfigurationReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.setAZConfigurationReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *DirectorService) SetNetworksConfiguration(arg1 json.RawMessage) error {
	fake.setNetworksConfigurationMutex.Lock()
	ret, specificReturn := fake.setNetworksConfigurationReturnsOnCall[len(fake.setNetworksConfigurationArgsForCall)]
	fake.setNetworksConfigurationArgsForCall = append(fake.setNetworksConfigurationArgsForCall, struct {
		arg1 json.RawMessage
	}{arg1})
	fake.recordInvocation("SetNetworksConfiguration", []interface{}{arg1})
	fake.setNetworksConfigurationMutex.Unlock()
	if fake.SetNetworksConfigurationStub != nil {
		return fake.SetNetworksConfigurationStub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.setNetworksConfigurationReturns.result1
}

func (fake *DirectorService) SetNetworksConfigurationCallCount() int {
	fake.setNetworksConfigurationMutex.RLock()
	defer fake.setNetworksConfigurationMutex.RUnlock()
	return len(fake.setNetworksConfigurationArgsForCall)
}

func (fake *DirectorService) SetNetworksConfigurationArgsForCall(i int) json.RawMessage {
	fake.setNetworksConfigurationMutex.RLock()
	defer fake.setNetworksConfigurationMutex.RUnlock()
	return fake.setNetworksConfigurationArgsForCall[i].arg1
}

func (fake *DirectorService) SetNetworksConfigurationReturns(result1 error) {
	fake.SetNetworksConfigurationStub = nil
	fake.setNetworksConfigurationReturns = struct {
		result1 error
	}{result1}
}

func (fake *DirectorService) SetNetworksConfigurationReturnsOnCall(i int, result1 error) {
	fake.SetNetworksConfigurationStub = nil
	if fake.setNetworksConfigurationReturnsOnCall == nil {
		fake.setNetworksConfigurationReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.setNetworksConfigurationReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *DirectorService) SetNetworkAndAZ(arg1 api.NetworkAndAZConfiguration) error {
	fake.setNetworkAndAZMutex.Lock()
	ret, specificReturn := fake.setNetworkAndAZReturnsOnCall[len(fake.setNetworkAndAZArgsForCall)]
	fake.setNetworkAndAZArgsForCall = append(fake.setNetworkAndAZArgsForCall, struct {
		arg1 api.NetworkAndAZConfiguration
	}{arg1})
	fake.recordInvocation("SetNetworkAndAZ", []interface{}{arg1})
	fake.setNetworkAndAZMutex.Unlock()
	if fake.SetNetworkAndAZStub != nil {
		return fake.SetNetworkAndAZStub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.setNetworkAndAZReturns.result1
}

func (fake *DirectorService) SetNetworkAndAZCallCount() int {
	fake.setNetworkAndAZMutex.RLock()
	defer fake.setNetworkAndAZMutex.RUnlock()
	return len(fake.setNetworkAndAZArgsForCall)
}

func (fake *DirectorService) SetNetworkAndAZArgsForCall(i int) api.NetworkAndAZConfiguration {
	fake.setNetworkAndAZMutex.RLock()
	defer fake.setNetworkAndAZMutex.RUnlock()
	return fake.setNetworkAndAZArgsForCall[i].arg1
}

func (fake *DirectorService) SetNetworkAndAZReturns(result1 error) {
	fake.SetNetworkAndAZStub = nil
	fake.setNetworkAndAZReturns = struct {
		result1 error
	}{result1}
}

func (fake *DirectorService) SetNetworkAndAZReturnsOnCall(i int, result1 error) {
	fake.SetNetworkAndAZStub = nil
	if fake.setNetworkAndAZReturnsOnCall == nil {
		fake.setNetworkAndAZReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.setNetworkAndAZReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *DirectorService) SetProperties(arg1 api.DirectorProperties) error {
	fake.setPropertiesMutex.Lock()
	ret, specificReturn := fake.setPropertiesReturnsOnCall[len(fake.setPropertiesArgsForCall)]
	fake.setPropertiesArgsForCall = append(fake.setPropertiesArgsForCall, struct {
		arg1 api.DirectorProperties
	}{arg1})
	fake.recordInvocation("SetProperties", []interface{}{arg1})
	fake.setPropertiesMutex.Unlock()
	if fake.SetPropertiesStub != nil {
		return fake.SetPropertiesStub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.setPropertiesReturns.result1
}

func (fake *DirectorService) SetPropertiesCallCount() int {
	fake.setPropertiesMutex.RLock()
	defer fake.setPropertiesMutex.RUnlock()
	return len(fake.setPropertiesArgsForCall)
}

func (fake *DirectorService) SetPropertiesArgsForCall(i int) api.DirectorProperties {
	fake.setPropertiesMutex.RLock()
	defer fake.setPropertiesMutex.RUnlock()
	return fake.setPropertiesArgsForCall[i].arg1
}

func (fake *DirectorService) SetPropertiesReturns(result1 error) {
	fake.SetPropertiesStub = nil
	fake.setPropertiesReturns = struct {
		result1 error
	}{result1}
}

func (fake *DirectorService) SetPropertiesReturnsOnCall(i int, result1 error) {
	fake.SetPropertiesStub = nil
	if fake.setPropertiesReturnsOnCall == nil {
		fake.setPropertiesReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.setPropertiesReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *DirectorService) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.setAZConfigurationMutex.RLock()
	defer fake.setAZConfigurationMutex.RUnlock()
	fake.setNetworksConfigurationMutex.RLock()
	defer fake.setNetworksConfigurationMutex.RUnlock()
	fake.setNetworkAndAZMutex.RLock()
	defer fake.setNetworkAndAZMutex.RUnlock()
	fake.setPropertiesMutex.RLock()
	defer fake.setPropertiesMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *DirectorService) recordInvocation(key string, args []interface{}) {
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
