package hypervctl

import (
	"fmt"
	"github.com/containers/libhvee/pkg/wmiext"
	"github.com/drtimf/wmi"
)

const (
	KvpOperationFailed    = 32768
	KvpAccessDenied       = 32769
	KvpNotSupported       = 32770
	KvpStatusUnknown      = 32771
	KvpTimeoutOcurred     = 32772
	KvpIllegalArgument    = 32773
	KvpSystemInUse        = 32774
	KvpInvalidState       = 32775
	KvpIncorrectDataType  = 32776
	KvpSystemNotAvailable = 32777
	KvpOutOfMemory        = 32778
	KvpNotFound           = 32779

	HyperVNamespace                = "root\\virtualization\\v2"
	VirtualSystemManagementService = "Msvm_VirtualSystemManagementService"
	KvpExchangeDataItemName        = "Msvm_KvpExchangeDataItem"
)

type VirtualMachineManager struct {
}

// NewVirtualMachineManager is a simple contructor for the VMM object
func NewVirtualMachineManager() VirtualMachineManager {
	return VirtualMachineManager{}
}

type KvpError struct {
	ErrorCode int
	message   string
}

func (k *KvpError) Error() string {
	return fmt.Sprintf("%s (%d)", k.message, k.ErrorCode)
}

func (*VirtualMachineManager) GetAll() ([]*VirtualMachine, error) {
	const wql = "Select * From Msvm_ComputerSystem"

	var service *wmi.Service
	var err error
	if service, err = wmi.NewLocalService(HyperVNamespace); err != nil {
		return [](*VirtualMachine){}, err
	}
	defer service.Close()

	var enum *wmi.Enum
	if enum, err = service.ExecQuery(wql); err != nil {
		return nil, err
	}
	defer enum.Close()

	var vms [](*VirtualMachine)
	for {
		vm := &VirtualMachine{}
		done, err := wmiext.NextObjectWithPath(enum, vm)
		if err != nil {
			return vms, err
		}
		if done {
			break
		}
		vms = append(vms, vm)
	}

	return vms, nil
}

func (*VirtualMachineManager) GetMachine(name string) (*VirtualMachine, error) {
	const wql = "Select * From Msvm_ComputerSystem Where ElementName='%s'"

	vm := &VirtualMachine{}
	var service *wmi.Service
	var err error

	if service, err = wmi.NewLocalService(HyperVNamespace); err != nil {
		return vm, err
	}
	defer service.Close()

	var enum *wmi.Enum
	if enum, err = service.ExecQuery(fmt.Sprintf(wql, name)); err != nil {
		return nil, err
	}
	defer enum.Close()

	done, err := wmiext.NextObjectWithPath(enum, vm)
	if err != nil {
		return vm, err
	}

	if done {
		return vm, fmt.Errorf("Could not find virtual machine %q", name)
	}

	return vm, nil
}

func translateError(source error, illegalSuggestion string) error {
	j, ok := source.(*wmiext.JobError)

	if !ok {
		return source
	}

	var message string
	switch j.ErrorCode {
	case KvpOperationFailed:
		message = "Operation failed"
	case KvpAccessDenied:
		message = "Access denied"
	case KvpNotSupported:
		message = "Not supported"
	case KvpStatusUnknown:
		message = "Status is unknown"
	case KvpTimeoutOcurred:
		message = "Timeout occurred"
	case KvpIllegalArgument:
		message = "Illegal argument (" + illegalSuggestion + ")"
	case KvpSystemInUse:
		message = "System is in use"
	case KvpInvalidState:
		message = "Invalid state for this operation"
	case KvpIncorrectDataType:
		message = "Incorrect data type"
	case KvpSystemNotAvailable:
		message = "System is not available"
	case KvpOutOfMemory:
		message = "Out of memory"
	case KvpNotFound:
		message = "Not found"
	default:
		return source
	}

	return &KvpError{j.ErrorCode, message}
}

func createItem(service *wmi.Service, key string, value string) string {
	item, err := wmiext.SpawnObject(service, KvpExchangeDataItemName)
	if err != nil {
		panic(err)
	}
	defer item.Close()

	item.Put("Name", key)
	item.Put("Data", value)
	item.Put("Source", 0)
	itemStr := wmiext.GetCimText(item)
	return itemStr
}
