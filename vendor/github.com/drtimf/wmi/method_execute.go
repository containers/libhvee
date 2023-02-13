// +build windows

package wmi

// MethodExecutor contains the context used for a fluent API to execute a WMI method
type MethodExecutor struct {
	err            error
	objectPath     string
	method         string
	service        *Service
	methodInstance *Instance
	inParam        *Instance
	outParam       *Instance
}

// BeginMethodExecute starts a fluent API call to a WMI method
func BeginMethodExecute(service *Service, instance *Instance, objectPath string, method string) *MethodExecutor {
	e := &MethodExecutor{method: method, objectPath: objectPath, service: service, methodInstance: instance}
	e.inParam, e.err = e.methodInstance.GetMethodParameters(e.method)
	return e
}

// Set provides a named parameter value to a WMI method call
func (e *MethodExecutor) Set(name string, value interface{}) *MethodExecutor {
	if e.err == nil {
		e.err = e.inParam.Put(name, value)
	}
	return e
}

// Execute performs the exectuion of the WMI method
func (e *MethodExecutor) Execute() *MethodExecutor {
	if e.err == nil {
		defer e.inParam.Close()

		if e.outParam, e.err = e.service.ExecMethod(e.objectPath, e.method, e.inParam); e.err == nil {
			e.err = checkRegistryReturn(e.outParam)
		}
	}

	return e
}

// Get obtains a named parameter result
func (e *MethodExecutor) Get(name string, value *interface{}) *MethodExecutor {
	if e.err == nil {
		var v interface{}
		if v, _, _, e.err = e.outParam.Get(name); e.err == nil {
			*value = v
		}
	}
	return e
}

// GetStringArray obtains a named parameter result as a string array
func (e *MethodExecutor) GetStringArray(name string, value *([]string)) *MethodExecutor {
	if e.err == nil {
		var v interface{}
		if v, _, _, e.err = e.outParam.Get(name); e.err == nil {
			if v == nil {
				*value = []string{}
			} else {
				*value = v.([]string)
			}
		}
	}
	return e
}

// End finishes the fluent API call providing an error if one occured at any stage
func (e *MethodExecutor) End() error {
	if e.err == nil {
		defer e.outParam.Close()
	}
	return e.err
}
