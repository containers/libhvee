// +build windows

package wmi

import (
	"fmt"
	"syscall"
	"unsafe"

	ole "github.com/go-ole/go-ole"
)

// WBEM_CONDITION_FLAG_TYPE contains flags used with the IWbemClassObject::GetNames method.
type WBEM_CONDITION_FLAG_TYPE uint32

const (
	WBEM_FLAG_ALWAYS                    WBEM_CONDITION_FLAG_TYPE = 0
	WBEM_FLAG_ONLY_IF_TRUE              WBEM_CONDITION_FLAG_TYPE = 0x1
	WBEM_FLAG_ONLY_IF_FALSE             WBEM_CONDITION_FLAG_TYPE = 0x2
	WBEM_FLAG_ONLY_IF_IDENTICAL         WBEM_CONDITION_FLAG_TYPE = 0x3
	WBEM_MASK_PRIMARY_CONDITION         WBEM_CONDITION_FLAG_TYPE = 0x3
	WBEM_FLAG_KEYS_ONLY                 WBEM_CONDITION_FLAG_TYPE = 0x4
	WBEM_FLAG_REFS_ONLY                 WBEM_CONDITION_FLAG_TYPE = 0x8
	WBEM_FLAG_LOCAL_ONLY                WBEM_CONDITION_FLAG_TYPE = 0x10
	WBEM_FLAG_PROPAGATED_ONLY           WBEM_CONDITION_FLAG_TYPE = 0x20
	WBEM_FLAG_SYSTEM_ONLY               WBEM_CONDITION_FLAG_TYPE = 0x30
	WBEM_FLAG_NONSYSTEM_ONLY            WBEM_CONDITION_FLAG_TYPE = 0x40
	WBEM_MASK_CONDITION_ORIGIN          WBEM_CONDITION_FLAG_TYPE = 0x70
	WBEM_FLAG_CLASS_OVERRIDES_ONLY      WBEM_CONDITION_FLAG_TYPE = 0x100
	WBEM_FLAG_CLASS_LOCAL_AND_OVERRIDES WBEM_CONDITION_FLAG_TYPE = 0x200
	WBEM_MASK_CLASS_CONDITION           WBEM_CONDITION_FLAG_TYPE = 0x300
)

// WBEM_FLAVOR_TYPE lists qualifier flavors
type WBEM_FLAVOR_TYPE uint32

const (
	WBEM_FLAVOR_DONT_PROPAGATE                  WBEM_FLAVOR_TYPE = 0
	WBEM_FLAVOR_FLAG_PROPAGATE_TO_INSTANCE      WBEM_FLAVOR_TYPE = 0x1
	WBEM_FLAVOR_FLAG_PROPAGATE_TO_DERIVED_CLASS WBEM_FLAVOR_TYPE = 0x2
	WBEM_FLAVOR_MASK_PROPAGATION                WBEM_FLAVOR_TYPE = 0xf
	WBEM_FLAVOR_OVERRIDABLE                     WBEM_FLAVOR_TYPE = 0
	WBEM_FLAVOR_NOT_OVERRIDABLE                 WBEM_FLAVOR_TYPE = 0x10
	WBEM_FLAVOR_MASK_PERMISSIONS                WBEM_FLAVOR_TYPE = 0x10
	WBEM_FLAVOR_ORIGIN_LOCAL                    WBEM_FLAVOR_TYPE = 0
	WBEM_FLAVOR_ORIGIN_PROPAGATED               WBEM_FLAVOR_TYPE = 0x20
	WBEM_FLAVOR_ORIGIN_SYSTEM                   WBEM_FLAVOR_TYPE = 0x40
	WBEM_FLAVOR_MASK_ORIGIN                     WBEM_FLAVOR_TYPE = 0x60
	WBEM_FLAVOR_NOT_AMENDED                     WBEM_FLAVOR_TYPE = 0
	WBEM_FLAVOR_AMENDED                         WBEM_FLAVOR_TYPE = 0x80
	WBEM_FLAVOR_MASK_AMENDED                    WBEM_FLAVOR_TYPE = 0x80
)

// IWbemClassObjectVtbl is the IWbemClassObject COM virtual table
type IWbemClassObjectVtbl struct {
	QueryInterface          uintptr
	AddRef                  uintptr
	Release                 uintptr
	GetQualifierSet         uintptr
	Get                     uintptr
	Put                     uintptr
	Delete                  uintptr
	GetNames                uintptr
	BeginEnumeration        uintptr
	Next                    uintptr
	EndEnumeration          uintptr
	GetPropertyQualifierSet uintptr
	Clone                   uintptr
	GetObjectText           uintptr
	SpawnDerivedClass       uintptr
	SpawnInstance           uintptr
	CompareTo               uintptr
	GetPropertyOrigin       uintptr
	InheritsFrom            uintptr
	GetMethod               uintptr
	PutMethod               uintptr
	DeleteMethod            uintptr
	BeginMethodEnumeration  uintptr
	NextMethod              uintptr
	EndMethodEnumeration    uintptr
	GetMethodQualifierSet   uintptr
	GetMethodOrigin         uintptr
}

// Instance of a WMI object
type Instance struct {
	classObject *ole.IUnknown
	classVTable *IWbemClassObjectVtbl
}

// Property name and value from a WMI object
type Property struct {
	Name  string
	Value interface{}
}

// ValueAsString returns the interface value as a string
func (p *Property) ValueAsString() (value string) {
	value = fmt.Sprintf("%v", p.Value)
	return
}

// NewInstance wraps an instance of a WMI object
func newInstance(classObject *ole.IUnknown) (instance *Instance) {
	instance = &Instance{
		classObject: classObject,
		classVTable: (*IWbemClassObjectVtbl)(unsafe.Pointer(classObject.RawVTable)),
	}

	return
}

// SpawnInstance creates a new instance of a WMI class. The current object must be a class definition obtained
// from WMI using GetObject or CreateClassEnum.
func (i *Instance) SpawnInstance() (instance *Instance, err error) {
	var hres uintptr
	var inst *ole.IUnknown

	hres, _, _ = syscall.Syscall6(i.classVTable.SpawnInstance, 3, // Call the IWbemClassObject::SpawnInstance method
		uintptr(unsafe.Pointer(i.classObject)),
		uintptr(0),                     // long lFlags,
		uintptr(unsafe.Pointer(&inst)), // IWbemClassObject **ppNewInstance
		uintptr(0),
		uintptr(0),
		uintptr(0))
	if FAILED(hres) {
		return nil, ole.NewError(hres)
	}

	return newInstance(inst), nil
}

// Get obtains a specified property value, if it exists.
func (i *Instance) Get(name string) (value interface{}, cimType CIMTYPE_ENUMERATION, flavor WBEM_FLAVOR_TYPE, err error) {
	var vtValue ole.VARIANT

	var nameUTF16 *uint16
	if nameUTF16, err = syscall.UTF16PtrFromString(name); err != nil {
		return
	}

	hres, _, _ := syscall.Syscall6(i.classVTable.Get, 6,
		uintptr(unsafe.Pointer(i.classObject)),
		uintptr(unsafe.Pointer(nameUTF16)), // LPCWSTR wszName
		uintptr(0),                         // long lFlags
		uintptr(unsafe.Pointer(&vtValue)),  // VARIANT *pVal
		uintptr(unsafe.Pointer(&cimType)),  // CIMTYPE *pType
		uintptr(unsafe.Pointer(&flavor)))   // long *plFlavor
	if FAILED(hres) {
		err = ole.NewError(hres)
		return
	}

	defer vtValue.Clear()
	value = VariantToValue(&vtValue)
	return
}

// Put sets a named property to a new value. This always overwrites the current value with a new one. When
// the instance is a CIM class definition, Put creates or updates the property value. When the instance
// is a CIM instance, Put updates a property value only. Put cannot create a property value.
func (i *Instance) Put(name string, value interface{}) (err error) {
	var hres uintptr
	vtValue := NewVariant(value)

	var nameUTF16 *uint16
	if nameUTF16, err = syscall.UTF16PtrFromString(name); err != nil {
		return
	}

	hres, _, _ = syscall.Syscall6(i.classVTable.Put, 5, // Call the IWbemClassObject::Put method
		uintptr(unsafe.Pointer(i.classObject)),
		uintptr(unsafe.Pointer(nameUTF16)), // LPCWSTR wszName
		uintptr(0),                         // long lFlags
		uintptr(unsafe.Pointer(&vtValue)),  // VARIANT *pVal
		uintptr(0),                         // CIMTYPE Type
		uintptr(0))
	if FAILED(hres) {
		return ole.NewError(hres)
	}

	vtValue.Clear()
	return
}

// GetNames retrieves the names of the properties in the WMI instance
func (i *Instance) GetNames() (names []string, err error) {
	var hres uintptr
	var classPropertyNames *ole.SafeArray

	hres, _, _ = syscall.Syscall6(i.classVTable.GetNames, 5, // Call the IWbemClassObject::GetNames method
		uintptr(unsafe.Pointer(i.classObject)),
		uintptr(0), //LPCWSTR wszQualifierName
		uintptr(WBEM_FLAG_ALWAYS|WBEM_FLAG_NONSYSTEM_ONLY), // long lFlags
		uintptr(0), // VARIANT *pQualifierVal
		uintptr(unsafe.Pointer(&classPropertyNames)), // SAFEARRAY **pNames
		uintptr(0))
	if FAILED(hres) {
		return nil, ole.NewError(hres)
	}

	// Convert the class property names into an ole.SafeArrayConversion object and enumerate
	// the class property names into a string array.
	safeClassPropertyNames := ole.SafeArrayConversion{Array: classPropertyNames}
	defer safeClassPropertyNames.Release()
	names = safeClassPropertyNames.ToStringArray()
	return
}

// BeginEnumeration resets an enumeration of instance properties back to the beginning of the enumeration.
// This must be called prior to the first call to Next to enumerate all of the properties on an object.
func (i *Instance) BeginEnumeration() (err error) {
	var hres uintptr

	hres, _, _ = syscall.Syscall6(i.classVTable.BeginEnumeration, 2, // Call the IWbemClassObject::BeginEnumeration method
		uintptr(unsafe.Pointer(i.classObject)),
		uintptr(WBEM_FLAG_ALWAYS|WBEM_FLAG_NONSYSTEM_ONLY), //long lFlags
		uintptr(0),
		uintptr(0),
		uintptr(0),
		uintptr(0))
	if FAILED(hres) {
		return ole.NewError(hres)
	}

	return nil
}

// NextAsVariant retrieves the next property in an enumeration as a variant that started with BeginEnumeration. This should be called repeatedly
// to enumerate all the properties until done returns true. If the enumeration is to be terminated early, then EndEnumeration should be called.
func (i *Instance) NextAsVariant() (done bool, name string, value *ole.VARIANT, cimType CIMTYPE_ENUMERATION, flavor WBEM_FLAVOR_TYPE, err error) {
	var hres uintptr
	var nameUTF16 *uint16
	var vtValue ole.VARIANT

	hres, _, _ = syscall.Syscall6(i.classVTable.Next, 6, // Call the IWbemClassObject::Next method
		uintptr(unsafe.Pointer(i.classObject)),
		uintptr(0),                          // long lFlags
		uintptr(unsafe.Pointer(&nameUTF16)), // BSTR *strName
		uintptr(unsafe.Pointer(&vtValue)),   // VARIANT *pVal
		uintptr(unsafe.Pointer(&cimType)),   // CIMTYPE *pType
		uintptr(unsafe.Pointer(&flavor)))    // long *plFlavor
	if FAILED(hres) {
		err = ole.NewError(hres)
		return
	}

	if hres == WBEM_S_NO_MORE_DATA {
		done = true
		return
	}

	defer ole.SysFreeString((*int16)(unsafe.Pointer(nameUTF16)))

	done = false
	name = ole.BstrToString(*(**uint16)(unsafe.Pointer(&nameUTF16)))
	value = &vtValue
	return
}

// Next retrieves the next property in an enumeration as a Go value that started with BeginEnumeration. This should be called repeatedly
// to enumerate all the properties until done returns true. If the enumeration is to be terminated early, then EndEnumeration should be called.
func (i *Instance) Next() (done bool, name string, value interface{}, cimType CIMTYPE_ENUMERATION, flavor WBEM_FLAVOR_TYPE, err error) {
	var vtValue *ole.VARIANT
	done, name, vtValue, cimType, flavor, err = i.NextAsVariant()

	if err == nil && !done {
		defer vtValue.Clear()
		value = VariantToValue(vtValue)
	}

	return
}

// EndEnumeration terminates an enumeration sequence started with BeginEnumeration. This call is not required,
// but it is recommended because it releases resources associated with the enumeration. However, the
// resources are deallocated automatically when the next enumeration is started or the object is released.
func (i *Instance) EndEnumeration() (err error) {
	var hres uintptr

	hres, _, _ = syscall.Syscall6(i.classVTable.EndEnumeration, 1, // Call the IWbemClassObject::EndEnumeration method
		uintptr(unsafe.Pointer(i.classObject)),
		uintptr(0),
		uintptr(0),
		uintptr(0),
		uintptr(0),
		uintptr(0))
	if FAILED(hres) {
		return ole.NewError(hres)
	}

	return
}

// BeginMethodEnumeration begins an enumeration of the methods available for the instance.
func (i *Instance) BeginMethodEnumeration() (err error) {
	var hres uintptr

	hres, _, _ = syscall.Syscall6(i.classVTable.BeginMethodEnumeration, 2, // Call the IWbemClassObject::BeginMethodEnumeration method
		uintptr(unsafe.Pointer(i.classObject)),
		uintptr(0), //long lEnumFlags
		uintptr(0),
		uintptr(0),
		uintptr(0),
		uintptr(0))
	if FAILED(hres) {
		return ole.NewError(hres)
	}

	return
}

// NextMethod retrieves the next method in a method enumeration sequence that starts with a call to BeginMethodEnumeration.
func (i *Instance) NextMethod() (done bool, name string, err error) {
	var hres uintptr
	var nameUTF16 *uint16

	hres, _, _ = syscall.Syscall6(i.classVTable.NextMethod, 5, // Call the IWbemClassObject::NextMethod method
		uintptr(unsafe.Pointer(i.classObject)),
		uintptr(0),                          //long lFlags
		uintptr(unsafe.Pointer(&nameUTF16)), //BSTR *pstrName
		uintptr(0),                          // IWbemClassObject **ppInSignature
		uintptr(0),                          // IWbemClassObject **ppOutSignature
		uintptr(0))
	if FAILED(hres) {
		err = ole.NewError(hres)
		return
	}

	if hres == WBEM_S_NO_MORE_DATA {
		done = true
		return
	}

	defer ole.SysFreeString((*int16)(unsafe.Pointer(nameUTF16)))

	done = false
	name = ole.BstrToString(*(**uint16)(unsafe.Pointer(&nameUTF16)))
	return
}

// EndMethodEnumeration terminates a method enumeration sequence started with BeginMethodEnumeration.
func (i *Instance) EndMethodEnumeration() (err error) {
	var hres uintptr
	hres, _, _ = syscall.Syscall6(i.classVTable.EndMethodEnumeration, 1, // Call the IWbemClassObject::EndMethodEnumeration method
		uintptr(unsafe.Pointer(i.classObject)),
		uintptr(0),
		uintptr(0),
		uintptr(0),
		uintptr(0),
		uintptr(0))
	if FAILED(hres) {
		return ole.NewError(hres)
	}

	return
}

// GetMethod obtains information about the requested method. This is only supported if the current instance is a
// CIM class definition. Method information is not available from instances which are CIM instances.
func (i *Instance) GetMethod(methodName string) (inSignature *Instance, outSignature *Instance, err error) {
	var hres uintptr
	var inSig, outSig *ole.IUnknown

	var methodNameUTF16 *uint16
	if methodNameUTF16, err = syscall.UTF16PtrFromString(methodName); err != nil {
		return
	}

	hres, _, _ = syscall.Syscall6(i.classVTable.GetMethod, 5, // Call the IWbemClassObject::GetMethod method
		uintptr(unsafe.Pointer(i.classObject)),
		uintptr(unsafe.Pointer(methodNameUTF16)), // LPCWSTR wszName
		uintptr(0),                               // long lFlags,
		uintptr(unsafe.Pointer(&inSig)),          // IWbemClassObject **ppInSignature,
		uintptr(unsafe.Pointer(&outSig)),         // IWbemClassObject **ppOutSignature
		uintptr(0))
	if FAILED(hres) {
		return nil, nil, ole.NewError(hres)
	}

	return newInstance(inSig), newInstance(outSig), nil
}

// GetMethodParameters obtains the input parameters of a method so they can be filled out for calling the method.
// This is a variation of GetMethod which only returns in input parameters.
func (i *Instance) GetMethodParameters(methodName string) (inParam *Instance, err error) {
	var hres uintptr
	var inSig *ole.IUnknown

	var methodNameUTF16 *uint16
	if methodNameUTF16, err = syscall.UTF16PtrFromString(methodName); err != nil {
		return
	}

	hres, _, _ = syscall.Syscall6(i.classVTable.GetMethod, 5, // Call the IWbemClassObject::GetMethod method
		uintptr(unsafe.Pointer(i.classObject)),
		uintptr(unsafe.Pointer(methodNameUTF16)), // LPCWSTR wszName
		uintptr(0),                               // long lFlags,
		uintptr(unsafe.Pointer(&inSig)),          // IWbemClassObject **ppInSignature,
		uintptr(0),                               // IWbemClassObject **ppOutSignature
		uintptr(0))
	if FAILED(hres) {
		return nil, ole.NewError(hres)
	}

	return newInstance(inSig), nil
}

// Close an instance
func (i *Instance) Close() {
	i.classObject.Release()
}

// GetPropertyAsString obtains a specificed property value as a string
func (i *Instance) GetPropertyAsString(name string) (value string, err error) {
	if v, _, _, err := i.Get(name); err != nil {
		return "", err
	} else {
		return fmt.Sprintf("%v", v), nil
	}
}

// GetProperties returns all the properties and their values for the WMI instance
func (i *Instance) GetProperties() (properties []Property, err error) {
	if err = i.BeginEnumeration(); err != nil {
		return
	}

	for done := false; !done; {
		var name string
		var value interface{}

		if done, name, value, _, _, err = i.Next(); err != nil {
			return
		}

		if !done {
			p := Property{
				Name:  name,
				Value: value,
			}

			properties = append(properties, p)
		}
	}

	i.EndEnumeration()
	return
}

// GetMethods obtain all the method names for the WMI instance
func (i *Instance) GetMethods() (methodNames []string, err error) {
	if err = i.BeginMethodEnumeration(); err != nil {
		return
	}

	for done := false; !done; {
		var name string
		if done, name, err = i.NextMethod(); err != nil {
			return
		}

		if !done {
			methodNames = append(methodNames, name)
		}
	}

	i.EndMethodEnumeration()

	return
}

// GetClassName obtains the class name of the WMI instance
func (i *Instance) GetClassName() (className string, err error) {
	return i.GetPropertyAsString(`__CLASS`)
}
