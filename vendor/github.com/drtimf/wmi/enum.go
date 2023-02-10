// +build windows

package wmi

import (
	"errors"
	"fmt"
	"reflect"
	"syscall"
	"unsafe"

	ole "github.com/go-ole/go-ole"
)

// WBEM_TIMEOUT_TYPE contains values used to specify the timeout for the IEnumWbemClassObject::Next method
type WBEM_TIMEOUT_TYPE uint32

const (
	WBEM_NO_WAIT  WBEM_TIMEOUT_TYPE = 0
	WBEM_INFINITE WBEM_TIMEOUT_TYPE = 0xFFFFFFFF
)

// IEnumWbemClassObjectVtbl is the IEnumWbemClassObject COM virtual table
type IEnumWbemClassObjectVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	Reset          uintptr
	Next           uintptr
	NextAsync      uintptr
	Clone          uintptr
	Skip           uintptr
}

// Enum wraps a WMI enumerator
type Enum struct {
	enumerator  *ole.IUnknown
	classVTable *IEnumWbemClassObjectVtbl
}

// NewEnum wraps a WMI enumerator
func newEnum(enumerator *ole.IUnknown) (i *Enum, err error) {
	return &Enum{
		enumerator:  enumerator,
		classVTable: (*IEnumWbemClassObjectVtbl)(unsafe.Pointer(enumerator.RawVTable)),
	}, nil
}

// Next obtains the next WMI instance from a WMI enumerator
func (e *Enum) Next() (instance *Instance, err error) {
	var hres uintptr
	var pclsObj *ole.IUnknown
	var uReturn uint32

	hres, _, _ = syscall.Syscall6(e.classVTable.Next, 5,
		uintptr(unsafe.Pointer(e.enumerator)), // Call the IEnumWbemClassObject::Next method
		uintptr(WBEM_INFINITE),
		uintptr(1),
		uintptr(unsafe.Pointer(&pclsObj)),
		uintptr(unsafe.Pointer(&uReturn)),
		uintptr(0))
	if FAILED(hres) {
		return nil, ole.NewError(hres)
	}

	if uReturn == 0 {
		if (hres != WBEM_S_NO_ERROR) && (hres != WBEM_S_FALSE) {
			return nil, fmt.Errorf("failed IEnumWbemClassObject::Next method, hres=%xh", hres)
		}
		return nil, nil
	}

	return newInstance(pclsObj), nil
}

// Close a WMI enumerator
func (e *Enum) Close() {
	e.enumerator.Release()
}

// NextObject obtains the next WMI instance from a WMI enumerator and maps it to a Go object
func (e *Enum) NextObject(dst interface{}) (done bool, err error) {
	dv := reflect.ValueOf(dst)
	if dv.Kind() != reflect.Ptr || dv.IsNil() {
		return false, errors.New("invalid desitnation type for mapping a WMI instance to an object")
	}

	dv = dv.Elem()

	var instance *Instance
	if instance, err = e.Next(); err != nil {
		return
	}

	if instance == nil {
		done = true
		return
	}

	defer instance.Close()
	done = false

	// Build a propert map of variants from the WMI object
	if err = instance.BeginEnumeration(); err != nil {
		return
	}

	propertyMap := make(map[string](*ole.VARIANT))

	for propDone := false; !propDone; {
		var name string
		var value *ole.VARIANT

		if propDone, name, value, _, _, err = instance.NextAsVariant(); err != nil {
			return
		}

		if !propDone && value != nil {
			propertyMap[name] = value
		}
	}

	// Ensure the proparty map is cleared
	defer func() {
		for _, v := range propertyMap {
			v.Clear()
		}
	}()

	instance.EndEnumeration()

	// Map the WMI proprties to the fields of the object
	t := reflect.TypeOf(dst).Elem()

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if variant, ok := propertyMap[f.Name]; ok {
			var val interface{}
			if val, err = VariantToGoType(variant, f.Type); err != nil {
				return
			}

			if val != nil {
				dv.Field(i).Set(reflect.ValueOf(val))
			}
		}
	}

	return
}
