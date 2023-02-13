// +build windows

// Some of this is derrived from invoke() in github.com/go-ole/go-ole/dispatch_windows.go
package wmi

import (
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"time"
	"unsafe"

	ole "github.com/go-ole/go-ole"
)

// NewVariant creates a variant from a go type
func NewVariant(v interface{}) (variant ole.VARIANT) {
	ole.VariantInit(&variant)

	switch vv := v.(type) {
	case bool:
		if vv {
			variant = ole.NewVariant(ole.VT_BOOL, 0xffff)
		} else {
			variant = ole.NewVariant(ole.VT_BOOL, 0)
		}
	case *bool:
		variant = ole.NewVariant(ole.VT_BOOL|ole.VT_BYREF, int64(uintptr(unsafe.Pointer(v.(*bool)))))
	case uint8:
		variant = ole.NewVariant(ole.VT_I1, int64(v.(uint8)))
	case *uint8:
		variant = ole.NewVariant(ole.VT_I1|ole.VT_BYREF, int64(uintptr(unsafe.Pointer(v.(*uint8)))))
	case int8:
		variant = ole.NewVariant(ole.VT_I1, int64(v.(int8)))
	case *int8:
		variant = ole.NewVariant(ole.VT_I1|ole.VT_BYREF, int64(uintptr(unsafe.Pointer(v.(*uint8)))))
	case int16:
		variant = ole.NewVariant(ole.VT_I2, int64(v.(int16)))
	case *int16:
		variant = ole.NewVariant(ole.VT_I2|ole.VT_BYREF, int64(uintptr(unsafe.Pointer(v.(*int16)))))
	case uint16:
		variant = ole.NewVariant(ole.VT_UI2, int64(v.(uint16)))
	case *uint16:
		variant = ole.NewVariant(ole.VT_UI2|ole.VT_BYREF, int64(uintptr(unsafe.Pointer(v.(*uint16)))))
	case int32:
		variant = ole.NewVariant(ole.VT_I4, int64(v.(int32)))
	case *int32:
		variant = ole.NewVariant(ole.VT_I4|ole.VT_BYREF, int64(uintptr(unsafe.Pointer(v.(*int32)))))
	case uint32:
		variant = ole.NewVariant(ole.VT_UI4, int64(v.(uint32)))
	case *uint32:
		variant = ole.NewVariant(ole.VT_UI4|ole.VT_BYREF, int64(uintptr(unsafe.Pointer(v.(*uint32)))))
	case int64:
		variant = ole.NewVariant(ole.VT_I8, int64(v.(int64)))
	case *int64:
		variant = ole.NewVariant(ole.VT_I8|ole.VT_BYREF, int64(uintptr(unsafe.Pointer(v.(*int64)))))
	case uint64:
		variant = ole.NewVariant(ole.VT_UI8, int64(uintptr(v.(uint64))))
	case *uint64:
		variant = ole.NewVariant(ole.VT_UI8|ole.VT_BYREF, int64(uintptr(unsafe.Pointer(v.(*uint64)))))
	case int:
		variant = ole.NewVariant(ole.VT_I4, int64(v.(int)))
	case *int:
		variant = ole.NewVariant(ole.VT_I4|ole.VT_BYREF, int64(uintptr(unsafe.Pointer(v.(*int)))))
	case uint:
		variant = ole.NewVariant(ole.VT_UI4, int64(v.(uint)))
	case *uint:
		variant = ole.NewVariant(ole.VT_UI4|ole.VT_BYREF, int64(uintptr(unsafe.Pointer(v.(*uint)))))
	case float32:
		variant = ole.NewVariant(ole.VT_R4, *(*int64)(unsafe.Pointer(&vv)))
	case *float32:
		variant = ole.NewVariant(ole.VT_R4|ole.VT_BYREF, int64(uintptr(unsafe.Pointer(v.(*float32)))))
	case float64:
		variant = ole.NewVariant(ole.VT_R8, *(*int64)(unsafe.Pointer(&vv)))
	case *float64:
		variant = ole.NewVariant(ole.VT_R8|ole.VT_BYREF, int64(uintptr(unsafe.Pointer(v.(*float64)))))
	case *big.Int:
		variant = ole.NewVariant(ole.VT_DECIMAL, v.(*big.Int).Int64())
	case string:
		variant = ole.NewVariant(ole.VT_BSTR, int64(uintptr(unsafe.Pointer(ole.SysAllocStringLen(v.(string))))))
	case *string:
		variant = ole.NewVariant(ole.VT_BSTR|ole.VT_BYREF, int64(uintptr(unsafe.Pointer(v.(*string)))))
	case time.Time:
		s := vv.Format("2006-01-02 15:04:05")
		variant = ole.NewVariant(ole.VT_BSTR, int64(uintptr(unsafe.Pointer(ole.SysAllocStringLen(s)))))
	case *time.Time:
		s := vv.Format("2006-01-02 15:04:05")
		variant = ole.NewVariant(ole.VT_BSTR|ole.VT_BYREF, int64(uintptr(unsafe.Pointer(&s))))
	case *ole.IDispatch:
		variant = ole.NewVariant(ole.VT_DISPATCH, int64(uintptr(unsafe.Pointer(v.(*ole.IDispatch)))))
	case **ole.IDispatch:
		variant = ole.NewVariant(ole.VT_DISPATCH|ole.VT_BYREF, int64(uintptr(unsafe.Pointer(v.(**ole.IDispatch)))))
	case nil:
		variant = ole.NewVariant(ole.VT_NULL, 0)
	case *ole.IUnknown:
		variant = ole.NewVariant(ole.VT_UNKNOWN, int64(uintptr(unsafe.Pointer(&(v.(*ole.IUnknown).RawVTable)))))
	case *Instance:
		variant = ole.NewVariant(ole.VT_UNKNOWN, int64(uintptr(unsafe.Pointer(&(v.(*Instance).classObject.RawVTable)))))
	default:
		panic(fmt.Sprintf("UNSUPPORTED for NewVariant: %T %v\n", v, v))
	}

	return
}

// VariantToValue extends the Value() method on an ole.VARIANT to deal with additional data types
func VariantToValue(variant *ole.VARIANT) (value interface{}) {
	if (variant.VT & ole.VT_ARRAY) != 0 {
		safeArrayConversion := ole.SafeArrayConversion{Array: *(**ole.SafeArray)(unsafe.Pointer(&variant.Val))}
		switch variant.VT {
		case ole.VT_ARRAY | ole.VT_BSTR:
			value = safeArrayConversion.ToStringArray()
		default:
			value = safeArrayConversion.ToValueArray()
		}
	} else {
		value = variant.Value()
	}
	return
}

func convertNumber(inputValue int64, outputKind reflect.Kind) (value interface{}, err error) {

	switch outputKind {
	case reflect.Bool:
		value = (inputValue != 0)
	case reflect.Int8:
		value = int8(inputValue)
	case reflect.Int16:
		value = int16(inputValue)
	case reflect.Int32:
		value = int32(inputValue)
	case reflect.Int64:
		value = int64(inputValue)
	case reflect.Uint8:
		value = uint8(inputValue)
	case reflect.Uint16:
		value = uint16(inputValue)
	case reflect.Uint32:
		value = uint32(inputValue)
	case reflect.Uint64:
		value = uint64(inputValue)
	default:
		err = fmt.Errorf("unsupported conversion of a number with value %v to %v", inputValue, outputKind)
	}

	return
}

func convertAnyNumber(inputValue interface{}, outputKind reflect.Kind) (value interface{}, err error) {

	switch vv := inputValue.(type) {
	case bool:
		if vv {
			return convertNumber(1, outputKind)
		} else {
			return convertNumber(0, outputKind)
		}
	case int8:
		return convertNumber(int64(vv), outputKind)
	case int16:
		return convertNumber(int64(vv), outputKind)
	case int32:
		return convertNumber(int64(vv), outputKind)
	case int64:
		return convertNumber(int64(vv), outputKind)
	case uint8:
		return convertNumber(int64(vv), outputKind)
	case uint16:
		return convertNumber(int64(vv), outputKind)
	case uint32:
		return convertNumber(int64(vv), outputKind)
	case uint64:
		return convertNumber(int64(vv), outputKind)
	}

	return
}

func converString(inputValue string, outputKind reflect.Kind) (value interface{}, err error) {

	var intOutput int64
	var uintOutput uint64
	var floatOutput float64

	switch outputKind {
	case reflect.String:
		value = inputValue
	case reflect.Bool:
		value, err = strconv.ParseBool(inputValue)
	case reflect.Int:
		intOutput, err = strconv.ParseInt(inputValue, 0, 64)
		value = int(intOutput)
	case reflect.Int8:
		intOutput, err = strconv.ParseInt(inputValue, 0, 8)
		value = int8(intOutput)
	case reflect.Int16:
		intOutput, err = strconv.ParseInt(inputValue, 0, 16)
		value = int16(intOutput)
	case reflect.Int32:
		intOutput, err = strconv.ParseInt(inputValue, 0, 32)
		value = int32(intOutput)
	case reflect.Int64:
		intOutput, err = strconv.ParseInt(inputValue, 0, 64)
		value = int64(intOutput)
	case reflect.Uint:
		uintOutput, err = strconv.ParseUint(inputValue, 0, 64)
		value = uint(uintOutput)
	case reflect.Uint8:
		uintOutput, err = strconv.ParseUint(inputValue, 0, 8)
		value = uint8(uintOutput)
	case reflect.Uint16:
		uintOutput, err = strconv.ParseUint(inputValue, 0, 16)
		value = uint16(uintOutput)
	case reflect.Uint32:
		uintOutput, err = strconv.ParseUint(inputValue, 0, 32)
		value = uint32(uintOutput)
	case reflect.Uint64:
		uintOutput, err = strconv.ParseUint(inputValue, 0, 64)
		value = uint64(uintOutput)
	case reflect.Float32:
		floatOutput, err = strconv.ParseFloat(inputValue, 32)
		value = float32(floatOutput)
	case reflect.Float64:
		floatOutput, err = strconv.ParseFloat(inputValue, 32)
		value = float64(floatOutput)
	default:
		err = fmt.Errorf("unsupported conversion of a string with value %v to %v", inputValue, outputKind)
	}

	return
}

func VariantToGoType(variant *ole.VARIANT, outputType reflect.Type) (value interface{}, err error) {
	switch variant.VT {
	case ole.VT_NULL:
		value = nil
	case ole.VT_BSTR:
		value, err = converString(variant.ToString(), outputType.Kind())
	case ole.VT_BOOL, ole.VT_UI1, ole.VT_UI2, ole.VT_UI4, ole.VT_UI8, ole.VT_I1, ole.VT_I2, ole.VT_I4, ole.VT_I8:
		value, err = convertNumber(variant.Val, outputType.Kind())
	case ole.VT_ARRAY | ole.VT_BSTR:
		safeArrayConversion := ole.SafeArrayConversion{Array: *(**ole.SafeArray)(unsafe.Pointer(&variant.Val))}
		value = safeArrayConversion.ToStringArray()
	case ole.VT_ARRAY | ole.VT_BOOL | ole.VT_ARRAY | ole.VT_UI1, ole.VT_ARRAY | ole.VT_UI2, ole.VT_ARRAY | ole.VT_UI4, ole.VT_ARRAY | ole.VT_UI8, ole.VT_ARRAY | ole.VT_I1, ole.VT_ARRAY | ole.VT_I2, ole.VT_ARRAY | ole.VT_I4, ole.VT_ARRAY | ole.VT_I8:
		safeArrayConversion := ole.SafeArrayConversion{Array: *(**ole.SafeArray)(unsafe.Pointer(&variant.Val))}
		valArray := safeArrayConversion.ToValueArray()
		valSlice := reflect.MakeSlice(reflect.SliceOf(outputType.Elem()), 0, len(valArray))
		for _, v := range valArray {
			var r interface{}
			if r, err = convertAnyNumber(v, outputType.Elem().Kind()); err != nil {
				break
			} else {
				valSlice = reflect.Append(valSlice, reflect.ValueOf(r))
			}
		}
		value = valSlice.Interface()

	default:
		err = fmt.Errorf("unsupported conversion of variant %v to %v", variant, outputType)
	}

	return
}
