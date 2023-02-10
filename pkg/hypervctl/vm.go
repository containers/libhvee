//go:build windows
// +build windows

package hypervctl

import (
	"fmt"
	"github.com/n1hility/hypervctl/pkg/wmiext"

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

type VirtualMachine struct {
	S__PATH                                  string `json:"-"`
	S__CLASS                                 string `json:"-"`
	InstanceID                               string
	Caption                                  string
	Description                              string
	ElementName                              string
	InstallDate                              string
	OperationalStatus                        []uint16
	StatusDescriptions                       []string
	Status                                   string
	HealthState                              uint16
	CommunicationStatus                      uint16
	DetailedStatus                           uint16
	OperatingStatus                          uint16
	PrimaryStatus                            uint16
	EnabledState                             uint16
	OtherEnabledState                        string
	RequestedState                           uint16
	EnabledDefault                           uint16
	TimeOfLastStateChange                    string
	AvailableRequestedStates                 []uint16
	TransitioningToState                     uint16
	CreationClassName                        string
	Name                                     string
	PrimaryOwnerName                         string
	PrimaryOwnerContact                      string
	Roles                                    []string
	NameFormat                               string
	OtherIdentifyingInfo                     []string
	IdentifyingDescriptions                  []string
	Dedicated                                []uint16
	OtherDedicatedDescriptions               []string
	ResetCapability                          uint16
	PowerManagementCapabilities              []uint16
	OnTimeInMilliseconds                     uint64
	ProcessID                                uint32
	TimeOfLastConfigurationChange            string
	NumberOfNumaNodes                        uint16
	ReplicationState                         uint16
	ReplicationHealth                        uint16
	ReplicationMode                          uint16
	FailedOverReplicationType                uint16
	LastReplicationType                      uint16
	LastApplicationConsistentReplicationTime string
	LastReplicationTime                      string
	LastSuccessfulBackupTime                 string
	EnhancedSessionModeState                 uint16
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

func (vm *VirtualMachine) AddKeyValuePair(key string, value string) error {
	return vm.kvpOperation("AddKvpItems", key, value, "key already exists?")
}

func (vm *VirtualMachine) ModifyKeyValuePair(key string, value string) error {
	return vm.kvpOperation("ModifyKvpItems", key, value, "key invalid?")
}

func (vm *VirtualMachine) PutKeyValuePair(key string, value string) error {
	err := vm.AddKeyValuePair(key, value)
	kvpError, ok := err.(*KvpError)
	if !ok || kvpError.ErrorCode != KvpIllegalArgument {
		return err
	}

	return vm.ModifyKeyValuePair(key, value)
}

func (vm *VirtualMachine) RemoveKeyValuePair(key string) error {
	return vm.kvpOperation("RemoveKvpItems", key, "", "key invalid?")
}

func (vm *VirtualMachine) kvpOperation(op string, key string, value string, illegalSuggestion string) error {
	var service *wmi.Service
	var vsms, job *wmi.Instance
	var err error

	if service, err = wmi.NewLocalService(HyperVNamespace); err != nil {
		return (err)
	}
	defer service.Close()

	vsms, err = wmiext.GetSingletonInstance(service, VirtualSystemManagementService)
	if err != nil {
		return err
	}
	defer vsms.Close()

	itemStr := createItem(service, key, value)

	execution := wmiext.BeginInvoke(service, vsms, op).
		Set("TargetSystem", vm.S__PATH).
		Set("DataItems", []string{itemStr}).
		Execute()

	if err := execution.Get("Job", &job).End(); err != nil {
		return fmt.Errorf("%s execution failed: %w", op, err)
	}

	err = translateError(wmiext.WaitJob(service, job), illegalSuggestion)
	defer job.Close()
	return err
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
