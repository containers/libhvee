// +build windows

package wmi

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"syscall"
	"unsafe"

	ole "github.com/go-ole/go-ole"
)

// Authentication Level Constants
const (
	RPC_C_AUTHN_LEVEL_DEFAULT       = 0
	RPC_C_AUTHN_LEVEL_NONE          = 1
	RPC_C_AUTHN_LEVEL_CONNECT       = 2
	RPC_C_AUTHN_LEVEL_CALL          = 3
	RPC_C_AUTHN_LEVEL_PKT           = 4
	RPC_C_AUTHN_LEVEL_PKT_INTEGRITY = 5
	RPC_C_AUTHN_LEVEL_PKT_PRIVACY   = 6
	RPC_C_AUTHN_WINNT               = 10
)

// EOLE_AUTHENTICATION_CAPABILITIES specifies various capabilities in CoInitializeSecurity
// and IClientSecurity::SetBlanket (or its helper function CoSetProxyBlanket).
type EOLE_AUTHENTICATION_CAPABILITIES uint32

const (
	EOAC_NONE              EOLE_AUTHENTICATION_CAPABILITIES = 0
	EOAC_MUTUAL_AUTH       EOLE_AUTHENTICATION_CAPABILITIES = 0x1
	EOAC_STATIC_CLOAKING   EOLE_AUTHENTICATION_CAPABILITIES = 0x20
	EOAC_DYNAMIC_CLOAKING  EOLE_AUTHENTICATION_CAPABILITIES = 0x40
	EOAC_ANY_AUTHORITY     EOLE_AUTHENTICATION_CAPABILITIES = 0x80
	EOAC_MAKE_FULLSIC      EOLE_AUTHENTICATION_CAPABILITIES = 0x100
	EOAC_DEFAULT           EOLE_AUTHENTICATION_CAPABILITIES = 0x800
	EOAC_SECURE_REFS       EOLE_AUTHENTICATION_CAPABILITIES = 0x2
	EOAC_ACCESS_CONTROL    EOLE_AUTHENTICATION_CAPABILITIES = 0x4
	EOAC_APPID             EOLE_AUTHENTICATION_CAPABILITIES = 0x8
	EOAC_DYNAMIC           EOLE_AUTHENTICATION_CAPABILITIES = 0x10
	EOAC_REQUIRE_FULLSIC   EOLE_AUTHENTICATION_CAPABILITIES = 0x200
	EOAC_AUTO_IMPERSONATE  EOLE_AUTHENTICATION_CAPABILITIES = 0x400
	EOAC_DISABLE_AAA       EOLE_AUTHENTICATION_CAPABILITIES = 0x1000
	EOAC_NO_CUSTOM_MARSHAL EOLE_AUTHENTICATION_CAPABILITIES = 0x2000
	EOAC_RESERVED1         EOLE_AUTHENTICATION_CAPABILITIES = 0x4000
)

// Authorization
const (
	RPC_C_AUTHZ_NONE    = 0
	RPC_C_AUTHZ_NAME    = 1
	RPC_C_AUTHZ_DCE     = 2
	RPC_C_AUTHZ_DEFAULT = 0xffffffff
)

// Impersonation Level Constants
const (
	RPC_C_IMP_LEVEL_DEFAULT     = 0
	RPC_C_IMP_LEVEL_ANONYMOUS   = 1
	RPC_C_IMP_LEVEL_IDENTIFY    = 2
	RPC_C_IMP_LEVEL_IMPERSONATE = 3
	RPC_C_IMP_LEVEL_DELEGATE    = 4
)

// The strings in this structure are in Unicode format.
const (
	SEC_WINNT_AUTH_IDENTITY_UNICODE = 2
)

// WBEM_GENERIC_FLAG_TYPE enumeration is used to indicate and update the type of the flag
type WBEM_GENERIC_FLAG_TYPE uint32

const (
	WBEM_FLAG_RETURN_WBEM_COMPLETE   WBEM_GENERIC_FLAG_TYPE = 0x0
	WBEM_FLAG_RETURN_IMMEDIATELY     WBEM_GENERIC_FLAG_TYPE = 0x10
	WBEM_FLAG_FORWARD_ONLY           WBEM_GENERIC_FLAG_TYPE = 0x20
	WBEM_FLAG_NO_ERROR_OBJECT        WBEM_GENERIC_FLAG_TYPE = 0x40
	WBEM_FLAG_SEND_STATUS            WBEM_GENERIC_FLAG_TYPE = 0x80
	WBEM_FLAG_ENSURE_LOCATABLE       WBEM_GENERIC_FLAG_TYPE = 0x100
	WBEM_FLAG_DIRECT_READ            WBEM_GENERIC_FLAG_TYPE = 0x200
	WBEM_MASK_RESERVED_FLAGS         WBEM_GENERIC_FLAG_TYPE = 0x1F000
	WBEM_FLAG_USE_AMENDED_QUALIFIERS WBEM_GENERIC_FLAG_TYPE = 0x20000
	WBEM_FLAG_STRONG_VALIDATION      WBEM_GENERIC_FLAG_TYPE = 0x100000
)

// IWbemServicesVtbl is the IWbemServices COM virtual table
type IWbemServicesVtbl struct {
	QueryInterface             uintptr
	AddRef                     uintptr
	Release                    uintptr
	OpenNamespace              uintptr
	CancelAsyncCall            uintptr
	QueryObjectSink            uintptr
	GetObject                  uintptr
	GetObjectAsync             uintptr
	PutClass                   uintptr
	PutClassAsync              uintptr
	DeleteClass                uintptr
	DeleteClassAsync           uintptr
	CreateClassEnum            uintptr
	CreateClassEnumAsync       uintptr
	PutInstance                uintptr
	PutInstanceAsync           uintptr
	DeleteInstance             uintptr
	DeleteInstanceAsync        uintptr
	CreateInstanceEnum         uintptr
	CreateInstanceEnumAsync    uintptr
	ExecQuery                  uintptr
	ExecQueryAsync             uintptr
	ExecNotificationQuery      uintptr
	ExecNotificationQueryAsync uintptr
	ExecMethod                 uintptr
	ExecMethodAsync            uintptr
}

// CoAuthIdentity represents the COAUTHITDENTIY structure
type CoAuthIdentity struct {
	User           *uint16
	UserLength     uint32
	Domain         *uint16
	DomainLength   uint32
	Password       *uint16
	PasswordLength uint32
	Flags          uint32
}

// WBEM_CONNECT_OPTIONS
const (
	WBEM_FLAG_CONNECT_REPOSITORY_ONLY uint32 = 0x40
	WBEM_FLAG_CONNECT_USE_MAX_WAIT    uint32 = 0x80
	WBEM_FLAG_CONNECT_PROVIDERS       uint32 = 0x100
)

// WBEM_QUERY_FLAG_TYPE
const (
	WBEM_FLAG_DEEP      uint32 = 0
	WBEM_FLAG_SHALLOW   uint32 = 1
	WBEM_FLAG_PROTOTYPE uint32 = 2
)

// CoSetProxyBlanket sets the authentication information that will be used to make calls on the specified proxy
func CoSetProxyBlanket(service *ole.IUnknown, identity *CoAuthIdentity) (err error) {
	var authnLevel uint32 = RPC_C_AUTHN_LEVEL_CALL
	if identity != nil {
		authnLevel = RPC_C_AUTHN_LEVEL_PKT_PRIVACY
	}

	hres, _, _ := procCoSetProxyBlanket.Call(
		uintptr(unsafe.Pointer(service)),     // IUnknown                 *pProxy,
		uintptr(RPC_C_AUTHN_WINNT),           // DWORD                    dwAuthnSvc,
		uintptr(RPC_C_AUTHZ_NONE),            // DWORD                    dwAuthzSvc,
		uintptr(0),                           // OLECHAR                  *pServerPrincName,
		uintptr(authnLevel),                  // DWORD                    dwAuthnLevel,
		uintptr(RPC_C_IMP_LEVEL_IMPERSONATE), // DWORD                    dwImpLevel,
		uintptr(unsafe.Pointer(identity)),    // RPC_AUTH_IDENTITY_HANDLE pAuthInfo,
		uintptr(EOAC_NONE))                   // DWORD                    dwCapabilities

	if FAILED(hres) {
		return ole.NewError(hres)
	}

	return nil
}

// Service holds the properties for a WMI connection
type Service struct {
	service     *ole.IUnknown
	classVTable *IWbemServicesVtbl
	identity    *CoAuthIdentity
}

// newService connects to a WMI service, local or remote using credentials
func newService(server string, namespace string, username string, password string) (s *Service, err error) {

	if wmiWbemLocator == nil {
		return nil, ole.NewError(WBEM_E_CRITICAL_ERROR)
	}

	var hres uintptr
	s = &Service{}

	var usernameUTF16 *uint16 = nil
	var passwordUTF16 *uint16 = nil

	if username != "" {
		var domain, user string
		if i := strings.Index(username, "\\"); i > 0 {
			domain = username[:i]
			user = username[i+1:]
		} else {
			domain = "."
			user = username
		}

		var userUTF16 *uint16
		if userUTF16, err = syscall.UTF16PtrFromString(user); err != nil {
			return
		}

		var domainUTF16 *uint16
		if domainUTF16, err = syscall.UTF16PtrFromString(domain); err != nil {
			return
		}

		if usernameUTF16, err = syscall.UTF16PtrFromString(domain + "\\" + user); err != nil {
			return
		}

		if passwordUTF16, err = syscall.UTF16PtrFromString(password); err != nil {
			return
		}

		s.identity = &CoAuthIdentity{
			Flags:          SEC_WINNT_AUTH_IDENTITY_UNICODE,
			Password:       passwordUTF16,
			PasswordLength: uint32(len(password)),
			Domain:         domainUTF16,
			DomainLength:   uint32(len(domain)),
			User:           userUTF16,
			UserLength:     uint32(len(user)),
		}
	}

	var networkResourceUTF16 *uint16
	if networkResourceUTF16, err = syscall.UTF16PtrFromString("\\\\" + server + "\\" + namespace); err != nil {
		return
	}

	myVTable := (*IWbemLocatorVtbl)(unsafe.Pointer(wmiWbemLocator.RawVTable))
	hres, _, _ = syscall.Syscall9(myVTable.ConnectServer, 9, // Call the IWbemLocator::ConnectServer method
		uintptr(unsafe.Pointer(wmiWbemLocator)),
		uintptr(unsafe.Pointer(networkResourceUTF16)), // const BSTR    strNetworkResource
		uintptr(unsafe.Pointer(usernameUTF16)),        // const BSTR    strUser
		uintptr(unsafe.Pointer(passwordUTF16)),        // const BSTR    strPassword
		uintptr(0),                                    // const BSTR    strLocale
		uintptr(WBEM_FLAG_CONNECT_USE_MAX_WAIT),       // long          lSecurityFlags
		uintptr(0),                                    // const BSTR    strAuthority
		uintptr(0),                                    // IWbemContext  *pCtx
		uintptr(unsafe.Pointer(&(s.service))))         // IWbemServices **ppNamespace

	if FAILED(hres) {
		return nil, ole.NewError(hres)
	}

	if err = CoSetProxyBlanket(s.service, s.identity); err != nil {
		return nil, err
	}

	s.classVTable = (*IWbemServicesVtbl)(unsafe.Pointer(s.service.RawVTable))

	return
}

// NewLocalService opens a connection to a local WMI service
func NewLocalService(namespace string) (s *Service, err error) {
	return newService(".", namespace, "", "")
}

// NewRemoteService a connection to a remote WMI service with a username and a password
func NewRemoteService(server string, namespace string, username string, password string) (s *Service, err error) {
	return newService(server, namespace, username, password)
}

// OpenNamespace creates a new service that has the specified child namespace as its operating context.
// All operations through the new service, such as class or instance creation, only affect that namespace. The namespace
// must be a child namespace of the current object through which this method is called.
func (s *Service) OpenNamespace(namespace string) (newService *Service, err error) {
	var hres uintptr
	newService = &Service{
		identity: s.identity,
	}

	var namespaceUTF16 *uint16
	if namespaceUTF16, err = syscall.UTF16PtrFromString(namespace); err != nil {
		return
	}

	hres, _, _ = syscall.Syscall6(s.classVTable.OpenNamespace, 6, // Call the IWbemServices::OpenNamespace method
		uintptr(unsafe.Pointer(s.service)),
		uintptr(unsafe.Pointer(namespaceUTF16)),        // const BSTR strNamespace
		uintptr(0),                                     // long lFlags
		uintptr(0),                                     // IWbemContext *pCtx
		uintptr(unsafe.Pointer(&(newService.service))), // IWbemServices **ppWorkingNamespace
		uintptr(0))                                     // IWbemCallResult **ppResult
	if FAILED(hres) {
		return nil, ole.NewError(hres)
	}

	if err = CoSetProxyBlanket(newService.service, newService.identity); err != nil {
		return nil, err
	}

	newService.classVTable = (*IWbemServicesVtbl)(unsafe.Pointer(newService.service.RawVTable))

	return
}

// CreateInstanceEnum obtains a WMI enumerator that returns the instances of a specified class.
func (s *Service) CreateInstanceEnum(className string) (e *Enum, err error) {
	var hres uintptr
	var pEnumerator *ole.IUnknown

	var classNameUTF16 *uint16
	if classNameUTF16, err = syscall.UTF16PtrFromString(className); err != nil {
		return
	}

	hres, _, _ = syscall.Syscall6(s.classVTable.CreateInstanceEnum, 5, // Call the IWbemServices::CreateInstanceEnum method
		uintptr(unsafe.Pointer(s.service)),
		uintptr(unsafe.Pointer(classNameUTF16)), // const BSTR strFilter
		uintptr(WBEM_FLAG_SHALLOW),              // long lFlags
		uintptr(0),                              // IWbemContext *pCtx
		uintptr(unsafe.Pointer(&pEnumerator)),   // IEnumWbemClassObject **ppEnum
		uintptr(0))
	if FAILED(hres) {
		return nil, ole.NewError(hres)
	}

	if err = CoSetProxyBlanket(pEnumerator, s.identity); err != nil {
		return nil, err
	}

	return newEnum(pEnumerator)
}

// ExecQuery executes a WMI Query Language (WQL) query and returns a WMI enumerator for the queried class instances
func (s *Service) ExecQuery(wqlQuery string) (e *Enum, err error) {
	var hres uintptr
	var pEnumerator *ole.IUnknown

	var queryLanguageUTF16 *uint16
	if queryLanguageUTF16, err = syscall.UTF16PtrFromString(`WQL`); err != nil {
		return
	}

	var wqlQueryUTF16 *uint16
	if wqlQueryUTF16, err = syscall.UTF16PtrFromString(wqlQuery); err != nil {
		return
	}

	hres, _, _ = syscall.Syscall6(s.classVTable.ExecQuery, 6, // Call the IWbemServices::ExecQuery method
		uintptr(unsafe.Pointer(s.service)),
		uintptr(unsafe.Pointer(queryLanguageUTF16)),                  // const BSTR strQueryLanguage
		uintptr(unsafe.Pointer(wqlQueryUTF16)),                       // const BSTR strQuery
		uintptr(WBEM_FLAG_FORWARD_ONLY|WBEM_FLAG_RETURN_IMMEDIATELY), // long lFlags
		uintptr(0),                            // IWbemContext *pCtx
		uintptr(unsafe.Pointer(&pEnumerator))) // IEnumWbemClassObject **ppEnum
	if FAILED(hres) {
		return nil, ole.NewError(hres)
	}

	if err = CoSetProxyBlanket(pEnumerator, s.identity); err != nil {
		return nil, err
	}

	return newEnum(pEnumerator)
}

// ExecMethod executes a method exported by a CIM object. The method call is forwarded to the appropriate provider where it executes.
// Information and status are returned to the caller, which blocks until the call is complete.
func (s *Service) ExecMethod(className string, methodName string, inParams *Instance) (outParam *Instance, err error) {
	var hres uintptr
	var outParams *ole.IUnknown

	var classNameUTF16 *uint16
	if classNameUTF16, err = syscall.UTF16PtrFromString(className); err != nil {
		return
	}

	var methodNameUTF16 *uint16
	if methodNameUTF16, err = syscall.UTF16PtrFromString(methodName); err != nil {
		return
	}

	hres, _, _ = syscall.Syscall9(s.classVTable.ExecMethod, 8, // Call the IWbemServices::ExecMethod method
		uintptr(unsafe.Pointer(s.service)),
		uintptr(unsafe.Pointer(classNameUTF16)),       // const BSTR strObjectPath
		uintptr(unsafe.Pointer(methodNameUTF16)),      // const BSTR strMethodName
		uintptr(0),                                    // long lFlags
		uintptr(0),                                    // IWbemContext *pCtx
		uintptr(unsafe.Pointer(inParams.classObject)), // IWbemClassObject *pInParams
		uintptr(unsafe.Pointer(&outParams)),           // IWbemClassObject **ppOutParams
		uintptr(0),                                    // IWbemCallResult **ppCallResult
		uintptr(0))
	if FAILED(hres) {
		return nil, ole.NewError(hres)
	}

	return newInstance(outParams), nil
}

// GetObject obtains a single WMI class or instance given its path
func (s *Service) GetObject(objectPath string) (instance *Instance, err error) {
	var hres uintptr
	var pObject *ole.IUnknown

	var objectPathUTF16 *uint16
	if objectPathUTF16, err = syscall.UTF16PtrFromString(objectPath); err != nil {
		return
	}

	hres, _, _ = syscall.Syscall6(s.classVTable.GetObject, 6, // Call the IWbemServices::GetObject method
		uintptr(unsafe.Pointer(s.service)),
		uintptr(unsafe.Pointer(objectPathUTF16)), // const BSTR       strObjectPath
		uintptr(WBEM_FLAG_RETURN_WBEM_COMPLETE),  // long             lFlags
		uintptr(0),                               // IWbemContext     *pCtx
		uintptr(unsafe.Pointer(&pObject)),        // IWbemClassObject **ppObject
		uintptr(0))                               // IWbemCallResult  **ppCallResult
	if FAILED(hres) {
		return nil, ole.NewError(hres)
	}

	return newInstance(pObject), nil
}

// Close a service connection
func (s *Service) Close() {
	s.service.Release()
}

// processEnumToObject enumerates a WMI enum into a struct or slice
func processEnumToObject(query string, enum *Enum, dst interface{}) (err error) {

	dt := reflect.TypeOf(dst)
	if dt.Kind() != reflect.Ptr {
		return errors.New("desitnation type for mapping a WMI instance to an object must be a pointer to a struct or slice")
	}

	dt = dt.Elem()
	dv := reflect.ValueOf(dst).Elem()

	isSlice := false
	if dt.Kind() == reflect.Slice {
		isSlice = true
		dv.Set(reflect.MakeSlice(dv.Type(), 0, 0))
		dt = dt.Elem()
	} else if dt.Kind() != reflect.Struct {
		return errors.New("desitnation pointer for mapping a WMI instance to an object must be a struct or slice")
	}

	if isSlice {
		for done := false; !done; {
			res := reflect.New(dt)
			if done, err = enum.NextObject(res.Interface()); err != nil {
				return
			}

			if !done {
				dv.Set(reflect.Append(dv, res.Elem()))
			}
		}
	} else {
		var done bool

		if done, err = enum.NextObject(dst); err != nil {
			return
		}

		if done {
			err = fmt.Errorf("expected at least one return WMI item from \"%s\" and recieved none", query)
		}
	}

	return
}

// Query executes a WMI Query Language (WQL) query and maps the results to a structure or slice of structures
// The destination must be a pointer to a struct:
//     var dst Win32_ComputerSystem
//     err = service.Query("SELECT * FROM Win32_ComputerSystem", &dst)
// Or a pointer to a slice:
//     var dst []CIM_DataFile
//     err = service.Query(`SELECT * FROM CIM_DataFile WHERE Drive = 'C:' AND Path = '\\'`, &dst)
func (s *Service) Query(query string, dst interface{}) (err error) {
	var enum *Enum
	if enum, err = s.ExecQuery(query); err != nil {
		return
	}
	defer enum.Close()

	return processEnumToObject(query, enum, dst)
}

// ClassInstances enumerates a WMI class of a given name and map the objects to a structure or slice of structures
func (s *Service) ClassInstances(className string, dst interface{}) (err error) {
	var enum *Enum
	if enum, err = s.CreateInstanceEnum(className); err != nil {
		return
	}
	defer enum.Close()

	return processEnumToObject(className, enum, dst)
}
