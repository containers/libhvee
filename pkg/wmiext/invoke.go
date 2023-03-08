//go:build windows
// +build windows

package wmiext

import (
	"reflect"

	"github.com/go-ole/go-ole"
)

type MethodExecutor struct {
	err      error
	path     string
	method   string
	service  *Service
	inParam  *Instance
	outParam *Instance
}

// In sets an input parameter for the method of this invocation, converting appropriately
func (e *MethodExecutor) In(name string, value interface{}) *MethodExecutor {
	if e.err == nil && e.inParam != nil {
		switch t := value.(type) {
		case *Instance:
			var ref bool
			if ref, e.err = e.inParam.IsReferenceProperty(name); e.err != nil {
				return e
			}
			if !ref {
				// Embedded Object
				break
			}
			if value, e.err = t.Path(); e.err != nil {
				return e
			}
		}

		e.err = e.inParam.Put(name, value)
	}

	return e
}

// Out sets the specified output parameter, and assigns the value parameter to the result.
// The value parameter must be a reference to the field that should be set.
func (e *MethodExecutor) Out(name string, value interface{}) *MethodExecutor {
	if e.err == nil && e.outParam != nil {
		var variant *ole.VARIANT
		var cimType CIMTYPE_ENUMERATION
		var result interface{}
		variant, cimType, _, e.err = e.outParam.GetAsVariant(name)
		defer variant.Clear()
		if e.err != nil || variant == nil {
			return e
		}
		if _, ok := value.(**Instance); ok && cimType == CIM_REFERENCE {
			path := variant.ToString()
			result, e.err = e.service.GetObject(path)
			if e.err != nil {
				return e
			}
		} else {
			target := reflect.ValueOf(value).Elem()
			result, e.err = convertToGoType(variant, target, target.Type())
			if e.err != nil {
				return e
			}
		}

		reflect.ValueOf(value).Elem().Set(reflect.ValueOf(result))
	}
	return e
}

// Execute executes the method after in parameters have been specified using In()
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
}

// End completes the method invocation and returns an error indicating the return
// code of the underlying method
func (e *MethodExecutor) End() error {
	e.cleanupInputs()

	if e.outParam != nil {
		e.outParam.Close()
		e.outParam = nil
	}

	return e.err
}

// Obtains the last error that occurred while building the invocation. Once
// an error has occurred, all future operations are treated as a no-op.
func (e *MethodExecutor) Error() error {
	return e.err
}
