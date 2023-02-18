//go:build windows
// +build windows

package hypervctl

import (
	"fmt"
	"strings"

	"github.com/drtimf/wmi"
	"github.com/n1hility/hypervctl/pkg/wmiext"
)

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

func (vm *VirtualMachine) Path() string {
	return vm.S__PATH
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

func (vm *VirtualMachine) GetKeyValuePairs() (map[string]string, error) {
	var service *wmi.Service
	var err error

	if service, err = wmi.NewLocalService(HyperVNamespace); err != nil {
		return nil, err
	}

	defer service.Close()

	i, err := wmiext.FindFirstRelatedInstance(service, vm.Path(), "Msvm_KvpExchangeComponent")
	if err != nil {
		return nil, err
	}

	defer i.Close()

	var path string
	path, err = i.GetPropertyAsString("__PATH")
	if err != nil {
		return nil, err

	}

	i, err = wmiext.FindFirstRelatedInstance(service, path, "Msvm_KvpExchangeComponentSettingData")
	if err != nil {
		return nil, err
	}
	defer i.Close()

	s, err := i.GetPropertyAsString("HostExchangeItems")
	if err != nil {
		return nil, err
	}

	return parseKvpMapXml(s)
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

	itemStr := createKvpItem(service, key, value)

	execution := wmiext.BeginInvoke(service, vsms, op).
		Set("TargetSystem", vm.Path()).
		Set("DataItems", []string{itemStr}).
		Execute()

	if err := execution.Get("Job", &job).End(); err != nil {
		return fmt.Errorf("%s execution failed: %w", op, err)
	}

	err = translateKvpError(wmiext.WaitJob(service, job), illegalSuggestion)
	defer job.Close()
	return err
}

func waitVMResult(res int32, service *wmi.Service, job *wmi.Instance) error {
	var err error

	if res == 4096 {
		err = wmiext.WaitJob(service, job)
		defer job.Close()
	}

	if err != nil {
		desc, _ := job.GetPropertyAsString("ErrorDescription")
		desc = strings.Replace(desc, "\n", " ", -1)
		return fmt.Errorf("Failed to define system: %w (%s)", err, desc)
	}

	return err
}
