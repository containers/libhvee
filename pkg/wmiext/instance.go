//go:build windows
// +build windows

package wmiext

import (
	"syscall"
	"unsafe"

	"github.com/drtimf/wmi"
	"github.com/go-ole/go-ole"
)

const WmiPathKey = "__PATH"

func RefetchObject(service *wmi.Service, instance *wmi.Instance) (*wmi.Instance, error) {
	path, err := ConvertToPath(instance)
	if err != nil {
		return instance, err
	}
	return service.GetObject(path)
}

func ConvertToPath(instance *wmi.Instance) (string, error) {
	ref, _, _, err := instance.Get(WmiPathKey)
	return ref.(string), err
}

func IsReferenceProperty(instance *wmi.Instance, name string) (bool, error) {
	_, cimType, _, err := instance.Get(name)
	return cimType == wmi.CIM_REFERENCE, err
}

func GetClassInstance(service *wmi.Service, obj *wmi.Instance) (*wmi.Instance, error) {
	name, err := obj.GetClassName()
	if err != nil {
		return nil, err
	}
	return service.GetObject(name)
}

func GetSingletonInstance(service *wmi.Service, className string) (*wmi.Instance, error) {
	var (
		enum     *wmi.Enum
		instance *wmi.Instance
		err      error
	)

	if enum, err = service.CreateInstanceEnum(className); err != nil {
		return nil, err
	}
	defer enum.Close()

	if instance, err = enum.Next(); err != nil {
		return nil, err
	}

	return instance, nil
}

func FindFirstObject(service *wmi.Service, wql string) (*wmi.Instance, error) {
	var enum *wmi.Enum
	var err error
	if enum, err = service.ExecQuery(wql); err != nil {
		return nil, err
	}

	defer enum.Close()
	return enum.Next()
}

func SpawnObject(service *wmi.Service, className string) (*wmi.Instance, error) {
	var class *wmi.Instance
	var err error
	if class, err = service.GetObject(className); err != nil {
		return nil, err
	}
	defer class.Close()

	return class.SpawnInstance()
}

// In order to implement the following 2 funcs we need the underlying
// class object for method invocation, which unfortunately isn't exported
func extractClassObj(instance *wmi.Instance) *ole.IUnknown {
	// First member of the struct
	return *(**ole.IUnknown)(unsafe.Pointer(instance))
}

// Alternative impl of instance.Put that supports direct passing of variants
// Once this is contributed back to wmi, this func can go
func InstancePut(i *wmi.Instance, name string, value interface{}) (err error) {
	var vtValue ole.VARIANT

	switch value.(type) {
	case ole.VARIANT:
		vtValue = value.(ole.VARIANT)
	case *ole.VARIANT:
		vtValue = *value.(*ole.VARIANT)
	default:
		vtValue = wmi.NewVariant(value)
	}

	var nameUTF16 *uint16
	if nameUTF16, err = syscall.UTF16PtrFromString(name); err != nil {
		return
	}

	classObj := extractClassObj(i)
	vTable := (*wmi.IWbemClassObjectVtbl)(unsafe.Pointer(classObj.RawVTable))
	ret, _, _ := syscall.SyscallN(vTable.Put, // IWbemClassObject::Put
		uintptr(unsafe.Pointer(classObj)),
		uintptr(unsafe.Pointer(nameUTF16)), // LPCWSTR wszName
		uintptr(0),                         // LONG lFlags
		uintptr(unsafe.Pointer(&vtValue)),  // VARIANT *pVal
		uintptr(0),                         // CIMTYPE Type
		uintptr(0))
	if ret != 0 {
		return ole.NewError(ret)
	}

	vtValue.Clear()
	return
}

func GetCimText(item *wmi.Instance) string {
	type wmiWbemTxtSrcVtable struct {
		QueryInterface uintptr
		AddRef         uintptr
		Release        uintptr
		GetTxt         uintptr
	}

	vTable := (*wmiWbemTxtSrcVtable)(unsafe.Pointer(wmiWbemTxtLocator.RawVTable))
	var retString *uint16
	_, _, _ = syscall.SyscallN(vTable.GetTxt, 0, 0, uintptr(unsafe.Pointer(extractClassObj(item))), 1, 0, uintptr(unsafe.Pointer(&retString)))

	itemStr := ole.BstrToString(retString)
	return itemStr
}
