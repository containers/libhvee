// +build windows

// A wrapper for local and remote Windows WMI at both low level calls to COM, and at a high level Go object mapping.
// There are a number of WMI library implementations around, but not many of them provide:
//    - Both local and remote access to the WMI provider
//    - A single session to execute many queries
//    - Low level access to the WMI API
//    - High level mapping of WMI objects to Go objects
//    - WMI method execution
// This presently only works on Windows. If there is ever a port of the Python Impacket to Go, it would be good to
// have this work on Linux and MacOS as well.
package wmi

import (
	"fmt"

	ole "github.com/go-ole/go-ole"
	"golang.org/x/sys/windows"
)

// Namespaces we use for WMI queries
const (
	RootCIMV2                   = `ROOT\CIMV2`
	RootMicrosoftWindowsStorage = `ROOT\Microsoft\Windows\Storage`
	RootMicrosoftSqlServer      = `ROOT\Microsoft\SqlServer`
	RootMSCluster               = `ROOT\MSCluster`
	RootWMI                     = `ROOT\WMI`
	RootVirtualization          = `ROOT\virtualization`
)

// HRESULT values
const (
	S_OK                     = 0
	S_FALSE                  = 1
	WBEM_S_NO_ERROR          = 0
	WBEM_S_FALSE             = 1
	WBEM_S_NO_MORE_DATA      = 0x40005
	WBEM_E_CRITICAL_ERROR    = 0x8004100A
	WBEM_E_NOT_SUPPORTED     = 0x8004100C
	WBEM_E_INVALID_NAMESPACE = 0x8004100E
	WBEM_E_INVALID_CLASS     = 0x80041010
)

// The CIMTYPE_ENUMERATION enumeration defines values that specify different CIM data types
type CIMTYPE_ENUMERATION uint32

const (
	CIM_ILLEGAL    CIMTYPE_ENUMERATION = 0xFFF
	CIM_EMPTY      CIMTYPE_ENUMERATION = 0
	CIM_SINT8      CIMTYPE_ENUMERATION = 16
	CIM_UINT8      CIMTYPE_ENUMERATION = 17
	CIM_SINT16     CIMTYPE_ENUMERATION = 2
	CIM_UINT16     CIMTYPE_ENUMERATION = 18
	CIM_SINT32     CIMTYPE_ENUMERATION = 3
	CIM_UINT32     CIMTYPE_ENUMERATION = 19
	CIM_SINT64     CIMTYPE_ENUMERATION = 20
	CIM_UINT64     CIMTYPE_ENUMERATION = 21
	CIM_REAL32     CIMTYPE_ENUMERATION = 4
	CIM_REAL64     CIMTYPE_ENUMERATION = 5
	CIM_BOOLEAN    CIMTYPE_ENUMERATION = 11
	CIM_STRING     CIMTYPE_ENUMERATION = 8
	CIM_DATETIME   CIMTYPE_ENUMERATION = 101
	CIM_REFERENCE  CIMTYPE_ENUMERATION = 102
	CIM_CHAR16     CIMTYPE_ENUMERATION = 103
	CIM_OBJECT     CIMTYPE_ENUMERATION = 13
	CIM_FLAG_ARRAY CIMTYPE_ENUMERATION = 0x2000
)

// IWbemLocatorVtbl is the IWbemLocator COM virtual table
type IWbemLocatorVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	ConnectServer  uintptr
}

var (
	// Load the ole32 DLL and bind methods
	ole32                    = windows.NewLazySystemDLL("ole32.dll")
	procCoInitializeSecurity = ole32.NewProc("CoInitializeSecurity")
	procCoSetProxyBlanket    = ole32.NewProc("CoSetProxyBlanket")

	// WMI Class and Interface GUIDs
	CLSID_WbemLocator    = ole.NewGUID("4590f811-1d3a-11d0-891f-00aa004b2e24")
	IID_IWbemLocator     = ole.NewGUID("dc12a687-737f-11cf-884d-00aa004b2e24")
	IID_IWbemClassObject = ole.NewGUID("dc12a681-737f-11cf-884d-00aa004b2e24")

	wmiWbemLocator *ole.IUnknown
	comInitialized bool
)

// Initialize the COM library
func init() {
	comInitialized = true
	err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
	if err != nil {
		comInitialized = false
		if oleCode, ok := err.(*ole.OleError); ok {
			switch oleCode.Code() {
			case S_OK, S_FALSE:
				comInitialized = true
			}
		}
		// Log error if unexpected failure occurs
		if !comInitialized {
			fmt.Printf("Unable to initialize COM, err=%v\n", err)
		}
	}

	// Set general COM security levels
	if comInitialized {
		hres, _, _ := procCoInitializeSecurity.Call(
			uintptr(0),
			uintptr(0xFFFFFFFF),                  // COM authentication
			uintptr(0),                           // Authentication services
			uintptr(0),                           // Reserved
			uintptr(RPC_C_AUTHN_LEVEL_DEFAULT),   // Default authentication
			uintptr(RPC_C_IMP_LEVEL_IMPERSONATE), // Default Impersonation
			uintptr(0),                           // Authentication info
			uintptr(EOAC_NONE),                   // Additional capabilities
			uintptr(0))                           // Reserved
		if FAILED(hres) {
			fmt.Printf("Unable to initialize COM security, err=%v\n", ole.NewError(hres))
		} else {
			// Obtain the initial locator to WMI
			wmiWbemLocator, err = ole.CreateInstance(CLSID_WbemLocator, IID_IWbemLocator)
			if err != nil {
				fmt.Printf("Unable to obtain the initial locator to WMI, err=%v\n", err)
				wmiWbemLocator = nil
			}
		}
	}
}

// Cleanup is an optional routine that should only be called when the process using the WMI package is exiting.
func Cleanup() {
	if wmiWbemLocator != nil {
		wmiWbemLocator.Release()
		wmiWbemLocator = nil
	}
	if comInitialized {
		ole.CoUninitialize()
	}
}

// SUCCEEDED function returns true if HRESULT succeeds, else false
func SUCCEEDED(hresult uintptr) bool {
	return int32(hresult) >= 0
}

// FAILED function returns true if HRESULT fails, else false
func FAILED(hresult uintptr) bool {
	return int32(hresult) < 0
}
