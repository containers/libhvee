//go:build windows
// +build windows

package wmiext

import (
	"fmt"
	"unsafe"

	"github.com/go-ole/go-ole"
)

func CreateStringArrayVariant(array []string) (ole.VARIANT, error) {
	safeByteArray, err := safeArrayFromStringSlice(array)
	if err != nil {
		return ole.VARIANT{}, err
	}
	arrayVariant := ole.NewVariant(ole.VT_ARRAY|ole.VT_BSTR, int64(uintptr(unsafe.Pointer(safeByteArray))))
	return arrayVariant, nil
}

// The following safearray routines are unfortunately not yet exported from go-ole,
// so replicate them for now
func safeArrayCreateVector(variantType ole.VT, lowerBound int32, length uint32) (safearray *ole.SafeArray, err error) {
	ret, _, _ := procSafeArrayCreateVector.Call(
		uintptr(variantType),
		uintptr(lowerBound),
		uintptr(length))

	if ret == 0 { // NULL return value
		err = fmt.Errorf("could not create safe array")
	}
	safearray = (*ole.SafeArray)(unsafe.Pointer(ret))
	return
}

func safeArrayDestroy(safearray *ole.SafeArray) (err error) {
	ret, _, _ := procSafeArrayDestroy.Call(uintptr(unsafe.Pointer(safearray)))

	if ret != 0 {
		return ole.NewError(ret)
	}

	return nil
}

func safeArrayPutElement(safearray *ole.SafeArray, index int64, element uintptr) (err error) {

	ret, _, _ := procSafeArrayPutElement.Call(
		uintptr(unsafe.Pointer(safearray)),
		uintptr(unsafe.Pointer(&index)),
		element)

	if ret != 0 {
		return ole.NewError(ret)
	}

	return nil
}

func safeArrayFromStringSlice(slice []string) (*ole.SafeArray, error) {
	array, err := safeArrayCreateVector(ole.VT_BSTR, 0, uint32(len(slice)))

	if err != nil {
		return nil, err
	}

	for i, v := range slice {
		err = safeArrayPutElement(array, int64(i), uintptr(unsafe.Pointer(ole.SysAllocStringLen(v))))
		if err != nil {
			_ = safeArrayDestroy(array)
			return nil, err
		}
	}
	return array, nil
}
