// +build windows

package wmi

import (
	ole "github.com/go-ole/go-ole"
)

type StdRegProvhDefKey uint32

const (
	HKEY_CLASSES_ROOT   StdRegProvhDefKey = 0x80000000
	HKEY_CURRENT_USER   StdRegProvhDefKey = 0x80000001
	HKEY_LOCAL_MACHINE  StdRegProvhDefKey = 0x80000002
	HKEY_USERS          StdRegProvhDefKey = 0x80000003
	HKEY_CURRENT_CONFIG StdRegProvhDefKey = 0x80000005
	HKEY_DYN_DATA       StdRegProvhDefKey = 0x80000006
)

type RegType int32

const (
	REG_SZ        RegType = 1
	REG_EXPAND_SZ RegType = 2
	REG_BINARY    RegType = 3
	REG_DWORD     RegType = 4
	REG_MULTI_SZ  RegType = 7
	REG_QWORD     RegType = 11
)

type Registry struct {
	service            *Service
	stdRegProvInstance *Instance
}

type RegistryValue struct {
	Name string
	Type RegType
}

func RegTypeToString(t RegType) string {
	switch t {
	case REG_SZ:
		return "REG_SZ"
	case REG_EXPAND_SZ:
		return "REG_EXPAND_SZ"
	case REG_BINARY:
		return "REG_BINARY"
	case REG_DWORD:
		return "REG_DWORD"
	case REG_MULTI_SZ:
		return "REG_MULTI_SZ"
	case REG_QWORD:
		return "REG_QWORD"
	}
	return "<UNKNOWN>"
}

func NewRegistry(service *Service) (wreg *Registry, err error) {
	var instance *Instance
	if instance, err = service.GetObject("StdRegProv"); err != nil {
		return
	}

	wreg = &Registry{
		service:            service,
		stdRegProvInstance: instance,
	}

	return
}

func checkRegistryReturn(param *Instance) (err error) {
	var n interface{}
	if n, _, _, err = param.Get("ReturnValue"); err != nil {
		return
	}

	if n.(int32) != 0 {
		err = ole.NewErrorWithDescription(uintptr(n.(int32)), "WMI registry call failed with error code")
	}

	return
}

// EnumKey uses WMI StdRegProv to enumerate registry key names
func (wreg *Registry) EnumKey(hkey StdRegProvhDefKey, subKeyName string) (names []string, err error) {
	err = BeginMethodExecute(wreg.service, wreg.stdRegProvInstance, "StdRegProv", "EnumKey").
		Set("hDefKey", int(hkey)).Set("sSubKeyName", subKeyName).
		Execute().
		GetStringArray("sNames", &names).
		End()
	return
}

// EnumValues uses WMI StdRegProv to enumerate registry key values of a subkey
func (wreg *Registry) EnumValues(hkey StdRegProvhDefKey, subKeyName string) (values []RegistryValue, err error) {
	var names []string
	var types interface{}
	if err = BeginMethodExecute(wreg.service, wreg.stdRegProvInstance, "StdRegProv", "EnumValues").
		Set("hDefKey", int(hkey)).Set("sSubKeyName", subKeyName).
		Execute().
		GetStringArray("sNames", &names).
		Get("Types", &types).
		End(); err == nil {

		values = make([]RegistryValue, len(names))
		for i, n := range names {
			values[i].Name = n
			values[i].Type = RegType((types.([]interface{}))[i].(int32))
		}
	}

	return
}

// GetStringValue gets a string from a registry value of type REG_SZ
func (wreg *Registry) GetStringValue(hkey StdRegProvhDefKey, subKeyName string, valueName string) (value string, err error) {
	var v interface{}
	err = BeginMethodExecute(wreg.service, wreg.stdRegProvInstance, "StdRegProv", "GetStringValue").
		Set("hDefKey", int(hkey)).Set("sSubKeyName", subKeyName).Set("sValueName", valueName).
		Execute().
		Get("sValue", &v).
		End()
	if err == nil {
		if v == nil {
			value = ""
		} else {
			value = v.(string)
		}
	}
	return
}

// GetExpandedStringValue gets a string from a registry value of type REG_EXPAND_SZ
func (wreg *Registry) GetExpandedStringValue(hkey StdRegProvhDefKey, subKeyName string, valueName string) (value string, err error) {
	var v interface{}
	err = BeginMethodExecute(wreg.service, wreg.stdRegProvInstance, "StdRegProv", "GetExpandedStringValue").
		Set("hDefKey", int(hkey)).Set("sSubKeyName", subKeyName).Set("sValueName", valueName).
		Execute().
		Get("sValue", &v).
		End()
	if err == nil {
		if v == nil {
			value = ""
		} else {
			value = v.(string)
		}
	}
	return
}

// GetBinaryValue gets a byte array from a registry value of type REG_BINARY
func (wreg *Registry) GetBinaryValue(hkey StdRegProvhDefKey, subKeyName string, valueName string) (value []uint8, err error) {
	var v interface{}
	err = BeginMethodExecute(wreg.service, wreg.stdRegProvInstance, "StdRegProv", "GetBinaryValue").
		Set("hDefKey", int(hkey)).Set("sSubKeyName", subKeyName).Set("sValueName", valueName).
		Execute().
		Get("uValue", &v).
		End()
	if err == nil {
		if v == nil {
			value = []uint8{}
		} else {
			value = make([]uint8, len(v.([]interface{})))
			for i, val := range v.([]interface{}) {
				value[i] = val.(uint8)
			}
		}
	}
	return
}

// GetDWORDValue gets an integer from a registry value of type REG_DWORD
func (wreg *Registry) GetDWORDValue(hkey StdRegProvhDefKey, subKeyName string, valueName string) (value int32, err error) {
	var v interface{}
	err = BeginMethodExecute(wreg.service, wreg.stdRegProvInstance, "StdRegProv", "GetDWORDValue").
		Set("hDefKey", int(hkey)).Set("sSubKeyName", subKeyName).Set("sValueName", valueName).
		Execute().
		Get("uValue", &v).
		End()
	if err == nil {
		if v == nil {
			value = -1
		} else {
			value = v.(int32)
		}
	}
	return
}

// GetMultiStringValue gets a string array from a registry value of type REG_MULTI_SZ
func (wreg *Registry) GetMultiStringValue(hkey StdRegProvhDefKey, subKeyName string, valueName string) (value []string, err error) {
	var v interface{}
	err = BeginMethodExecute(wreg.service, wreg.stdRegProvInstance, "StdRegProv", "GetMultiStringValue").
		Set("hDefKey", int(hkey)).Set("sSubKeyName", subKeyName).Set("sValueName", valueName).
		Execute().
		Get("sValue", &v).
		End()
	if err == nil {
		if v == nil {
			value = []string{}
		} else {
			value = v.([]string)
		}
	}
	return
}

// GetQWORDValue gets a string version of a uint64 from a registry value of type REG_QWORD (REVISIT: the variant is a BSTR rather than VT_UI8)
func (wreg *Registry) GetQWORDValue(hkey StdRegProvhDefKey, subKeyName string, valueName string) (value string, err error) {
	var v interface{}
	err = BeginMethodExecute(wreg.service, wreg.stdRegProvInstance, "StdRegProv", "GetQWORDValue").
		Set("hDefKey", int(hkey)).Set("sSubKeyName", subKeyName).Set("sValueName", valueName).
		Execute().
		Get("uValue", &v).
		End()
	if err == nil {
		if v == nil {
			value = ""
		} else {
			value = v.(string)
		}
	}
	return
}

// GetValue returns the value of a registry key
func (wreg *Registry) GetValue(hkey StdRegProvhDefKey, subKeyName string, valueType RegType, valueName string) (value interface{}, err error) {
	switch valueType {
	case REG_SZ:
		value, err = wreg.GetStringValue(hkey, subKeyName, valueName)
	case REG_EXPAND_SZ:
		value, err = wreg.GetExpandedStringValue(hkey, subKeyName, valueName)
	case REG_BINARY:
		value, err = wreg.GetBinaryValue(hkey, subKeyName, valueName)
	case REG_DWORD:
		value, err = wreg.GetDWORDValue(hkey, subKeyName, valueName)
	case REG_MULTI_SZ:
		value, err = wreg.GetMultiStringValue(hkey, subKeyName, valueName)
	case REG_QWORD:
		value, err = wreg.GetQWORDValue(hkey, subKeyName, valueName)
	}

	return
}

// Close releases the registry context
func (wreg *Registry) Close() {
	wreg.stdRegProvInstance.Close()
}
