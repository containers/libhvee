# wmi
A wrapper for local and remote Windows WMI at both low level calls to COM, and at a high level Go object mapping.

There are a number of WMI library implementations around, but not many of them provide:
* Both local and remote access to the WMI provider
* A single session to execute many queries
* Low level access to the WMI API
* High level mapping of WMI objects to Go objects
* WMI method execution

This presently only works on Windows.  If there is ever a port of the Python Impacket to Go, it would be good to have this work on Linux and MacOS as well.

[![Build Status](https://travis-ci.com/drtimf/wmi.svg?branch=main)](https://travis-ci.com/drtimf/wmi)
[![GoDoc](https://pkg.go.dev/badge/github.com/drtimf/wmi)](https://pkg.go.dev/github.com/drtimf/wmi)

## Examples

### A Simple High-Level Example to Query Win32_ComputerSystem

```go
package main

import (
	"fmt"

	"github.com/drtimf/wmi"
)

func main() {
	var service *wmi.Service
	var err error

	if service, err = wmi.NewLocalService(wmi.RootCIMV2); err != nil {
		panic(err)
	}

	defer service.Close()

	var computerSystem wmi.Win32_ComputerSystem
	if err = service.Query("SELECT * FROM Win32_ComputerSystem", &computerSystem); err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", computerSystem)
}
```

### A Low-Level Example to Query Win32_NetworkAdapter

```go
package main

import (
	"fmt"

	"github.com/drtimf/wmi"
)

func main() {
	var service *wmi.Service
	var err error

	if service, err = wmi.NewLocalService(wmi.RootCIMV2); err != nil {
		panic(err)
	}

	defer service.Close()

	var enum *wmi.Enum
	if enum, err = service.ExecQuery(`SELECT InterfaceIndex, Manufacturer, MACAddress, Name FROM Win32_NetworkAdapter`); err != nil {
		panic(err)
	}

	defer enum.Close()

	for {
		var instance *wmi.Instance
		if instance, err = enum.Next(); err != nil {
			panic(err)
		}

		if instance == nil {
			break
		}

		defer instance.Close()

		var val interface{}
		var interfaceIndex int32
		var manufacturer, MACAddress, name string

		if val, _, _, err = instance.Get("InterfaceIndex"); err != nil {
			panic(err)
		}

		if val != nil {
			interfaceIndex = val.(int32)
		}

		if val, _, _, err = instance.Get("Manufacturer"); err != nil {
			panic(err)
		}

		if val != nil {
			manufacturer = val.(string)
		}

		if val, _, _, err = instance.Get("MACAddress"); err != nil {
			panic(err)
		}

		if val != nil {
			MACAddress = val.(string)
		}

		if val, _, _, err = instance.Get("Name"); err != nil {
			panic(err)
		}

		if val != nil {
			name = val.(string)
		}

		fmt.Printf("%6d %-25s%20s\t%s\n", interfaceIndex, manufacturer, MACAddress, name)
	}
}
```

## Usage

Objects in this API include a Close() method which should be called when the object is no longer required.  This is important as it invokes the COM Release() to free the resources memory.

There are three high level objects:
* Service: A connection to either a local or remote WMI service
* Enum: An enumerator of WMI instances
* Instance: An instance of a WMI class

### Open a Connection to a WMI Service and Namespace

In each case a new *wmi.Service is created which can be used to obtain class instances and execute queries.

Open a connection to a local WMI service
``` go
func NewLocalService(namespace string) (s *Service, err error)
````

Open a connection to a remote WMI service with a username and a password:
``` go
func NewRemoteService(server string, namespace string, username string, password string) (s *Service, err error)
```

Create a new service that has the specified child namespace as its operating context.  All operations through the new service, such as class or instance creation, only affect that namespace. The namespace must be a child namespace of the current object through which this method is called.
``` go
func (s *Service) OpenNamespace(namespace string) (newService *Service, err error)
```

### Get a Single WMI Object

Obtain a single WMI class or instance given its path:
``` go
func (s *Service) GetObject(objectPath string) (instance *Instance, err error)
```

### Query WMI Objects

Obtain a WMI enumerator that returns the instances of a specified class:
``` go
func (s *Service) CreateInstanceEnum(className string) (e *Enum, err error)
```

Enumerate a WMI class of a given name and map the objects to a structure or slice of structures:
``` go
func (s *Service) ClassInstances(className string, dst interface{}) (err error)
```

Execute a WMI Query Language (WQL) query and return a WMI enumerator for the queried class instances:
``` go
func (s *Service) ExecQuery(wqlQuery string) (e *Enum, err error)
```

Execute a WMI Query Language (WQL) query and map the results to a structure or slice of structures:
``` go
func (s *Service) Query(query string, dst interface{}) (err error)
```
The destination must be a pointer to a struct:
```go
var dst Win32_ComputerSystem
err = service.Query("SELECT * FROM Win32_ComputerSystem", &dst)
```
Or a pointer to a slice:
``` go
var dst []CIM_DataFile
err = service.Query(`SELECT * FROM CIM_DataFile WHERE Drive = 'C:' AND Path = '\\'`, &dst)
```

### Execute a WMI Method

Execute a method exported by a CIM object:
``` go
func (s *Service) ExecMethod(className string, methodName string, inParams *Instance) (outParam *Instance, err error)
```
The method call is forwarded to the appropriate provider where it executes.  Information and status are returned to the caller, which blocks until the call is complete.

### Enumerate WMI objects

Obtain the next WMI instance from a WMI enumerator:
``` go
func (e *Enum) Next() (instance *Instance, err error)
```

Obtain the next WMI instance from a WMI enumerator and map it to a Go object:
``` go
func (e *Enum) NextObject(dst interface{}) (done bool, err error)
```

### Create Class Instances

Create a new instance of a WMI class:
``` go
func (i *Instance) SpawnInstance() (instance *Instance, err error)
```
The current object must be a class definition obtained from WMI using GetObject or CreateClassEnum.

### Get Properties of a WMI Instance

Obtain the class name of the WMI instance:
``` go
func (i *Instance) GetClassName() (className string, err error)
```

Retrieve the names of the properties in the WMI instance:
``` go
func (i *Instance) GetNames() (names []string, err error)
```

Obtain a specified property value, if it exists:
``` go
func (i *Instance) Get(name string) (value interface{}, cimType CIMTYPE_ENUMERATION, flavor WBEM_FLAVOR_TYPE, err error)
```

Obtain a specificed property value as a string:
``` go
func (i *Instance) GetPropertyAsString(name string) (value string, err error)
```

Reset an enumeration of instance properties back to the beginning of the enumeration:
``` go
func (i *Instance) BeginEnumeration() (err error)
```
This must be called prior to the first call to Next to enumerate all of the properties on an object.

Retrieves the next property in an enumeration as a variant that started with BeginEnumeration:
``` go
func (i *Instance) NextAsVariant() (done bool, name string, value *ole.VARIANT, cimType CIMTYPE_ENUMERATION, flavor WBEM_FLAVOR_TYPE, err error)
```
This should be called repeatedly to enumerate all the properties until done returns true. If the enumeration is to be terminated early, then EndEnumeration should be called.

Retrieve the next property in an enumeration as a Go value that started with BeginEnumeration:
``` go
func (i *Instance) Next() (done bool, name string, value interface{}, cimType CIMTYPE_ENUMERATION, flavor WBEM_FLAVOR_TYPE, err error)
```

Terminate an enumeration sequence started with BeginEnumeration:
``` go
func (i *Instance) EndEnumeration() (err error)
```
This call is not required, but it is recommended because it releases resources associated with the enumeration. However, the resources are deallocated automatically when the next enumeration is started or the object is released.

Return all the properties and their values for the WMI instance
``` go
func (i *Instance) GetProperties() (properties []Property, err error)
```

### Set Properties for a WMI Instance

Set a named property to a new value:
``` go
func (i *Instance) Put(name string, value interface{}) (err error)
```
This always overwrites the current value with a new one. When the instance is a CIM class definition, Put creates or updates the property value. When the instance is a CIM instance, Put updates a property value only. Put cannot create a property value.


### Methods of a WMI Instance

Begin an enumeration of the methods available for the instance:
``` go
func (i *Instance) BeginMethodEnumeration() (err error)
```

Retrieve the next method in a method enumeration sequence that starts with a call to BeginMethodEnumeration:
``` go
func (i *Instance) NextMethod() (done bool, name string, err error)
```

Terminate a method enumeration sequence started with BeginMethodEnumeration:
``` go
func (i *Instance) EndMethodEnumeration() (err error)
```

Obtain all the method names for the WMI instance:
``` go
func (i *Instance) GetMethods() (methodNames []string, err error)
```

Obtain information about the requested method:
``` go
func (i *Instance) GetMethod(methodName string) (inSignature *Instance, outSignature *Instance, err error)
```
This is only supported if the current instance is a CIM class definition. Method information is not available from instances which are CIM instances.

Obtain the input parameters of a method so they can be filled out for calling the method:
``` go
func (i *Instance) GetMethodParameters(methodName string) (inParam *Instance, err error)
```
This is a variation of GetMethod which only returns in input parameters.

