//go:build windows
// +build windows

package wmiext

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/drtimf/wmi"
	"github.com/go-ole/go-ole"
)

const (
	WmiPathKey = "__PATH"
)

var (
	WindowsEpoch = time.Date(1601, 1, 1, 0, 0, 0, 0, time.UTC)
)

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

func GetObjectAsObject(service *wmi.Service, objPath string, target interface{}) error {
	instance, err := service.GetObject(objPath)
	if err != nil {
		return err
	}
	defer instance.Close()

	return InstanceGetAll(instance, target)
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

func FindFirstInstance(service *wmi.Service, wql string) (*wmi.Instance, error) {
	var enum *wmi.Enum
	var err error
	if enum, err = service.ExecQuery(wql); err != nil {
		return nil, err
	}
	defer enum.Close()

	instance, err := enum.Next()
	if err != nil {
		return nil, err
	}

	if instance == nil {
		return nil, errors.New("No results found.")
	}

	return instance, nil
}

func FindFirstRelatedInstance(service *wmi.Service, objPath string, className string) (*wmi.Instance, error) {
	wql := fmt.Sprintf("ASSOCIATORS OF {%s} WHERE ResultClass = %s", objPath, className)
	return FindFirstInstance(service, wql)
}

func FindFirstObject(service *wmi.Service, wql string, target interface{}) error {
	var enum *wmi.Enum
	var err error
	if enum, err = service.ExecQuery(wql); err != nil {
		return err
	}
	defer enum.Close()

	done, err := NextObjectWithPath(enum, target)
	if err != nil {
		return err
	}

	if done {
		return errors.New("no results found")
	}

	return nil
}

func CreateInstance(service *wmi.Service, className string, src interface{}) (*wmi.Instance, error) {
	instance, err := SpawnInstance(service, className)
	if err != nil {
		return nil, err
	}

	return instance, InstancePutAll(instance, src)
}

func SpawnInstance(service *wmi.Service, className string) (*wmi.Instance, error) {
	var class *wmi.Instance
	var err error
	if class, err = service.GetObject(className); err != nil {
		return nil, err
	}
	defer class.Close()

	return class.SpawnInstance()
}

func CloneInstance(instance *wmi.Instance) (*wmi.Instance, error) {
	classObj := extractClassObj(instance)
	vTable := (*wmi.IWbemClassObjectVtbl)(unsafe.Pointer(classObj.RawVTable))
	var cloned *ole.IUnknown

	ret, _, _ := syscall.SyscallN(vTable.Clone, // IWbemClassObject::Put
		uintptr(unsafe.Pointer(classObj)),
		uintptr(unsafe.Pointer(&cloned)))
	if ret != 0 {
		return nil, ole.NewError(ret)
	}

	// Copy the full struct to copy over the vtable, then change the first
	// pointer to the new cloned handle
	copy := *instance
	ref := (**ole.IUnknown)(unsafe.Pointer(&copy))
	*ref = cloned

	return &copy, nil
}

// In order to implement the following 2 funcs we need the underlying
// class object for method invocation, which unfortunately isn't exported
func extractClassObj(instance *wmi.Instance) *ole.IUnknown {
	// First member of the struct
	return *(**ole.IUnknown)(unsafe.Pointer(instance))
}

func InstancePutAll(instance *wmi.Instance, src interface{}) error {
	val := reflect.ValueOf(src)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return errors.New("not a struct or pointer to struct")
	}

	props, err := instance.GetProperties()
	if err != nil {
		return err
	}

	propMap := make(map[string]struct{})
	for _, prop := range props {
		propMap[prop.Name] = struct{}{}
	}

	return instancePutAllTraverse(instance, val, propMap)
}

func instancePutAllTraverse(instance *wmi.Instance, val reflect.Value, propMap map[string]struct{}) error {
	for i := 0; i < val.NumField(); i++ {
		fieldVal := val.Field(i)
		fieldType := val.Type().Field(i)

		if fieldType.Type.Kind() == reflect.Struct && fieldType.Anonymous {
			if err := instancePutAllTraverse(instance, fieldVal, propMap); err != nil {
				return err
			}
			continue
		}
		if strings.HasPrefix(fieldType.Name, "S__") {
			continue
		}

		if !fieldType.IsExported() {
			continue
		}

		if _, exists := propMap[fieldType.Name]; !exists {
			continue
		}

		if fieldVal.Kind() == reflect.String && fieldVal.Len() == 0 {
			continue
		}

		if err := InstancePut(instance, fieldType.Name, fieldVal.Interface()); err != nil {
			return err
		}
	}

	return nil
}

// Alternative impl of instance.Put that supports direct passing of variants
// Once this is contributed back to wmi, this func can go
func InstancePut(i *wmi.Instance, name string, value interface{}) (err error) {
	var vtValue ole.VARIANT

	switch cast := value.(type) {
	case ole.VARIANT:
		vtValue = cast
	case *ole.VARIANT:
		vtValue = *cast
	default:
		vtValue, err = NewAutomationVariant(value)
		if err != nil {
			return err
		}
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

	_ = vtValue.Clear()
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

func GetPropertyAsUint(instance *wmi.Instance, name string) (uint, error) {
	val, _, _, err := instance.Get(name)
	if err != nil {
		return 0, err
	}

	switch ret := val.(type) {
	case int:
		return uint(ret), nil
	case int8:
		return uint(ret), nil
	case int16:
		return uint(ret), nil
	case int32:
		return uint(ret), nil
	case int64:
		return uint(ret), nil
	case uint:
		return ret, nil
	case uint8:
		return uint(ret), nil
	case uint16:
		return uint(ret), nil
	case uint32:
		return uint(ret), nil
	case uint64:
		return uint(ret), nil
	case string:
		parse, err := strconv.ParseUint(ret, 10, 64)
		return uint(parse), err
	default:
		return 0, fmt.Errorf("Type conversion from %T on param %s not supported", val, name)
	}
}

func InstanceGetAll(instance *wmi.Instance, target interface{}) error {
	elem := reflect.ValueOf(target)
	if elem.Kind() != reflect.Ptr || elem.IsNil() {
		return errors.New("invalid destination type for mapping a WMI instance to an object")
	}

	// deref pointer
	elem = elem.Elem()
	var err error

	if err = enumerateWithSystem(instance); err != nil {
		return err
	}

	properties := make(map[string](*ole.VARIANT))

	for {
		var name string
		var value *ole.VARIANT
		var done bool

		if done, name, value, _, _, err = instance.NextAsVariant(); err != nil {
			return err
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
			_ = v.Clear()
		}
	}()

	_ = instance.EndEnumeration()

	return instanceGetAllPopulate(elem, elem.Type(), properties)
}

func instanceGetAllPopulate(elem reflect.Value, elemType reflect.Type, properties map[string]*ole.VARIANT) error {
	var err error

	for i := 0; i < elemType.NumField(); i++ {
		fieldType := elemType.Field(i)
		fieldVal := elem.Field(i)

		if !fieldType.IsExported() {
			continue
		}

		if fieldType.Type.Kind() == reflect.Struct && fieldType.Anonymous {
			if err := instanceGetAllPopulate(fieldVal, fieldType.Type, properties); err != nil {
				return err
			}
			continue
		}

		fieldName := fieldType.Name
		if strings.HasPrefix(fieldName, "S__") {
			fieldName = fieldName[1:]
		}
		if variant, ok := properties[fieldName]; ok {
			var val interface{}
			if val, err = convertToGoType(variant, fieldVal, fieldType.Type); err != nil {
				return err
			}

			if val != nil {
				fieldVal.Set(reflect.ValueOf(val))
			}
		}
	}

	return nil
}

func convertToGoType(variant *ole.VARIANT, outputValue reflect.Value, outputType reflect.Type) (value interface{}, err error) {
	switch outputValue.Interface().(type) {
	case time.Time:
		return convertDataTimeToTime(variant)
	case *time.Time:
		x, err := convertDataTimeToTime(variant)
		return &x, err
	}

	return wmi.VariantToGoType(variant, outputType)
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

// Automation variants do not follow the OLE rules, instead they use the following mapping:
// sint8	VT_I2	Signed 8-bit integer.
// sint16	VT_I2	Signed 16-bit integer.
// sint32	VT_I4	Signed 32-bit integer.
// sint64	VT_BSTR	Signed 64-bit integer in string form. This type follows hexadecimal or decimal format
//                  according to the American National Standards Institute (ANSI) C rules.
// real32	VT_R4	4-byte floating-point value that follows the Institute of Electrical and Electronics
//                  Engineers, Inc. (IEEE) standard.
// real64	VT_R8	8-byte floating-point value that follows the IEEE standard.
// uint8	VT_UI1	Unsigned 8-bit integer.
// uint16	VT_I4	Unsigned 16-bit integer.
// uint32	VT_I4	Unsigned 32-bit integer.
// uint64	VT_BSTR	Unsigned 64-bit integer in string form. This type follows hexadecimal or decimal format
//                  according to ANSI C rules.

func NewAutomationVariant(value interface{}) (ole.VARIANT, error) {
	switch cast := value.(type) {
	case bool:
		if cast {
			return ole.NewVariant(ole.VT_BOOL, 0xffff), nil
		} else {
			return ole.NewVariant(ole.VT_BOOL, 0), nil
		}
	case int8:
		return ole.NewVariant(ole.VT_I2, int64(cast)), nil
	case []int8:
		return CreateNumericArrayVariant(cast, ole.VT_I2)
	case int16:
		return ole.NewVariant(ole.VT_I2, int64(cast)), nil
	case []int16:
		return CreateNumericArrayVariant(cast, ole.VT_I2)
	case int32:
		return ole.NewVariant(ole.VT_I4, int64(cast)), nil
	case []int32:
		return CreateNumericArrayVariant(cast, ole.VT_I4)
	case int64:
		s := fmt.Sprintf("%d", cast)
		return ole.NewVariant(ole.VT_BSTR, int64(uintptr(unsafe.Pointer(ole.SysAllocStringLen(s))))), nil
	case []int64:
		strings := make([]string, len(cast))
		for i, num := range cast {
			strings[i] = fmt.Sprintf("%d", num)
		}
		return CreateStringArrayVariant(strings)
	case float32:
		return ole.NewVariant(ole.VT_R4, int64(math.Float32bits(cast))), nil
	case float64:
		return ole.NewVariant(ole.VT_R8, int64(math.Float64bits(cast))), nil
	case uint8:
		return ole.NewVariant(ole.VT_UI1, int64(cast)), nil
	case []uint8:
		return CreateNumericArrayVariant(cast, ole.VT_UI1)
	case uint16:
		return ole.NewVariant(ole.VT_I4, int64(cast)), nil
	case []uint16:
		return CreateNumericArrayVariant(cast, ole.VT_I4)
	case uint32:
		return ole.NewVariant(ole.VT_I4, int64(cast)), nil
	case []uint32:
		return CreateNumericArrayVariant(cast, ole.VT_I4)
	case uint64:
		s := fmt.Sprintf("%d", cast)
		return ole.NewVariant(ole.VT_BSTR, int64(uintptr(unsafe.Pointer(ole.SysAllocStringLen(s))))), nil
	case []uint64:
		strings := make([]string, len(cast))
		for i, num := range cast {
			strings[i] = fmt.Sprintf("%d", num)
		}
		return CreateStringArrayVariant(strings)

	// Assume 32 bit for generic (u)ints
	case int:
		return ole.NewVariant(ole.VT_I4, int64(cast)), nil
	case uint:
		return ole.NewVariant(ole.VT_I4, int64(cast)), nil
	case []int:
		return CreateNumericArrayVariant(cast, ole.VT_I4)
	case []uint:
		return CreateNumericArrayVariant(cast, ole.VT_I4)

	case string:
		return ole.NewVariant(ole.VT_BSTR, int64(uintptr(unsafe.Pointer(ole.SysAllocStringLen(value.(string)))))), nil
	case []string:
		if len(cast) == 0 {
			return ole.NewVariant(ole.VT_NULL, 0), nil
		}
		return CreateStringArrayVariant(cast)

	case time.Time:
		return convertTimeToDataTime(&cast), nil
	case *time.Time:
		return convertTimeToDataTime(cast), nil
	case nil:
		return ole.NewVariant(ole.VT_NULL, 0), nil
	case *ole.IUnknown:
		if cast == nil {
			return ole.NewVariant(ole.VT_NULL, 0), nil
		}
		return ole.NewVariant(ole.VT_UNKNOWN, int64(uintptr(unsafe.Pointer(&(value.(*ole.IUnknown).RawVTable))))), nil
	case *wmi.Instance:
		if cast == nil {
			return ole.NewVariant(ole.VT_NULL, 0), nil
		}
		return ole.NewVariant(ole.VT_UNKNOWN, int64(uintptr(unsafe.Pointer(&(extractClassObj(value.(*wmi.Instance)).RawVTable))))), nil
	default:
		return ole.VARIANT{}, fmt.Errorf("unsupported type for automation variants %T", value)
	}
}

func convertTimeToDataTime(time *time.Time) ole.VARIANT {
	if time == nil || !time.After(WindowsEpoch) {
		return ole.NewVariant(ole.VT_NULL, 0)
	}
	_, offset := time.Zone()
	// convert to minutes
	offset /= 60
	//yyyymmddHHMMSS.mmmmmmsUUU
	s := fmt.Sprintf("%s%+04d", time.Format("20060102150405.000000"), offset)
	return ole.NewVariant(ole.VT_BSTR, int64(uintptr(unsafe.Pointer(ole.SysAllocStringLen(s)))))
}

func convertDataTimeToTime(variant *ole.VARIANT) (time.Time, error) {
	var zeroTime = time.Time{}
	var dateTime string

	switch variant.VT {
	case ole.VT_BSTR:
		dateTime = variant.ToString()
	case ole.VT_NULL:
		return zeroTime, nil
	default:
		return zeroTime, errors.New("Variant not compatible with dateTime field")
	}

	dLen := len(dateTime)
	if dLen < 5 {
		return zeroTime, errors.New("Invalid datetime string")
	}

	if strings.HasPrefix(dateTime, "00000000000000") {
		// Zero time
		return zeroTime, nil
	}

	zoneStart := dLen - 4
	var zoneMinutes int64
	var err error
	if dateTime[zoneStart] == ':' {
		// interval ends in :000 since not TZ based
		zoneMinutes = 0
	} else {
		zoneSuffix := dateTime[zoneStart:dLen]
		zoneMinutes, err = strconv.ParseInt(zoneSuffix, 10, 0)
		if err != nil {
			return zeroTime, errors.New("Invalid datetime string, zone did not parse")
		}
	}

	timePortion := dateTime[0:zoneStart]
	timePortion = fmt.Sprintf("%s%+03d%02d", timePortion, zoneMinutes/60, abs(int(zoneMinutes%60)))
	return time.Parse("20060102150405.000000-0700", timePortion)
}

func abs(num int) int {
	if num < 0 {
		return -num
	}

	return num
}
