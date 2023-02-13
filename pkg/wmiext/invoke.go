//go:build windows
// +build windows

package wmiext

import (
	"reflect"

	"github.com/drtimf/wmi"
	"github.com/go-ole/go-ole"
)

type MethodExecutor struct {
	err      error
	path     string
	method   string
	service  *wmi.Service
	inParam  *wmi.Instance
	outParam *wmi.Instance
	vars     []*ole.VARIANT
}

func (e *MethodExecutor) Set(name string, value interface{}) *MethodExecutor {
	if e.err == nil && e.inParam != nil {
		switch t := value.(type) {
		case []string:
			var array ole.VARIANT
			if array, e.err = CreateStringArrayVariant(t); e.err != nil {
				return e
			}
			e.vars = append(e.vars, &array)
			value = array
		case *wmi.Instance:
			var ref bool
			if ref, e.err = IsReferenceProperty(e.inParam, name); e.err != nil {
				return e
			}
			if !ref {
				// Embedded Object
				break
			}
			if value, e.err = ConvertToPath(t); e.err != nil {
				return e
			}
		}

		e.err = InstancePut(e.inParam, name, value)
	}

	return e
}

func (e *MethodExecutor) Get(name string, value interface{}) *MethodExecutor {
	if e.err == nil && e.outParam != nil {
		var result interface{}
		var cimType wmi.CIMTYPE_ENUMERATION
		result, cimType, _, e.err = e.outParam.Get(name)
		if e.err != nil {
			return e
		}
		if _, ok := value.(**wmi.Instance); ok && cimType == wmi.CIM_REFERENCE {
			result, e.err = e.service.GetObject(result.(string))
			if e.err != nil {
				return e
			}
		}

		reflect.ValueOf(value).Elem().Set(reflect.ValueOf(result))
	}
	return e
}

func (e *MethodExecutor) Execute() *MethodExecutor {
	defer e.cleanupInputs()

	if e.err == nil {
		e.outParam, e.err = e.service.ExecMethod(e.path, e.method, e.inParam)
	}

	return e
}

func (e *MethodExecutor) cleanupInputs() {
	if e.inParam != nil {
		e.inParam.Close()
		e.inParam = nil
	}
	e.vars = nil
}

func (e *MethodExecutor) End() error {
	e.cleanupInputs()

	if e.outParam != nil {
		e.outParam.Close()
		e.outParam = nil
	}

	return e.err
}

func (e *MethodExecutor) Error() error {
	return e.err
}

func BeginInvoke(service *wmi.Service, obj *wmi.Instance, method string) *MethodExecutor {
	objPath, err := ConvertToPath(obj)
	if err != nil {
		return &MethodExecutor{err: err}
	}

	var class, inParam *wmi.Instance
	if class, err = GetClassInstance(service, obj); err == nil {
		inParam, err = class.GetMethodParameters(method)
		class.Close()
	}

	return &MethodExecutor{method: method, path: objPath, service: service, inParam: inParam, err: err}
}
