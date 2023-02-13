//go:build windows
// +build windows

package wmiext

import (
	"errors"
	"reflect"
	"strings"
	"syscall"
	"unsafe"

	"github.com/drtimf/wmi"
	"github.com/go-ole/go-ole"
)

// Variant of Enum.NextObject that also sets object and class paths if present
func NextObjectWithPath(enum *wmi.Enum, target interface{}) (bool, error) {
	var err error
	elem := reflect.ValueOf(target)
	if elem.Kind() != reflect.Ptr || elem.IsNil() {
		return false, errors.New("invalid destination type for mapping a WMI instance to an object")
	}
	elem = elem.Elem()

	var instance *wmi.Instance
	if instance, err = enum.Next(); err != nil {
		return false, err
	}

	if instance == nil {
		return true, nil
	}

	defer instance.Close()

	if err = enumerateWithSystem(instance); err != nil {
		return false, err
	}

	properties := make(map[string](*ole.VARIANT))

	for {
		var name string
		var value *ole.VARIANT
		var done bool

		if done, name, value, _, _, err = instance.NextAsVariant(); err != nil {
			return false, err
		}

		if done {
			break
		}

		if value != nil {
			properties[name] = value
		}
	}

	defer func() {
		for _, v := range properties {
			v.Clear()
		}
	}()

	_ = instance.EndEnumeration()

	t := reflect.TypeOf(target).Elem()

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		fieldName := f.Name
		if strings.HasPrefix(fieldName, "S__") {
			fieldName = fieldName[1:]
		}
		if variant, ok := properties[fieldName]; ok {
			var val interface{}
			if val, err = wmi.VariantToGoType(variant, f.Type); err != nil {
				return false, err
			}

			if val != nil {
				elem.Field(i).Set(reflect.ValueOf(val))
			}
		}
	}

	return false, nil
}

func enumerateWithSystem(instance *wmi.Instance) (err error) {
	classObj := extractClassObj(instance)
	vTable := (*wmi.IWbemClassObjectVtbl)(unsafe.Pointer(classObj.RawVTable))

	result, _, _ := syscall.SyscallN(vTable.BeginEnumeration, uintptr(unsafe.Pointer(classObj)), 0)
	if result != 0 {
		return ole.NewError(result)
	}

	return nil
}
