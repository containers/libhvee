//go:build windows
// +build windows

package hypervctl

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/containers/libhvee/pkg/kvp/ginsu"
	"github.com/containers/libhvee/pkg/wmiext"
	"github.com/drtimf/wmi"
)

// delete this when close to being done
var (
	ErrNotImplemented = errors.New("function not implemented")
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
func (vm *VirtualMachine) SplitAndAddIgnition(keyPrefix string, ignRdr *bytes.Reader) error {
	parts, err := ginsu.Dice(ignRdr)
	if err != nil {
		return err
	}
	for idx, val := range parts {
		key := fmt.Sprintf("%s%d", keyPrefix, idx)
		if err := vm.AddKeyValuePair(key, val); err != nil {
			return err
		}
	}
	return nil
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
		return fmt.Errorf("failed to define system: %w (%s)", err, desc)
	}

	return err
}

func (vm *VirtualMachine) Stop() error {
	if !Enabled.equal(vm.EnabledState) {
		return ErrMachineNotRunning
	}
	var (
		err error
		job *wmi.Instance
		res int32
		srv *wmi.Service
	)
	if srv, err = wmi.NewLocalService(HyperVNamespace); err != nil {
		return err
	}
	wmiInst, err := wmiext.FindFirstRelatedInstance(srv, vm.Path(), "Msvm_ShutdownComponent")
	if err != nil {
		return err
	}
	// https://learn.microsoft.com/en-us/windows/win32/hyperv_v2/msvm-shutdowncomponent-initiateshutdown
	err = wmiext.BeginInvoke(srv, wmiInst, "InitiateShutdown").
		Set("Reason", "User requested").
		Set("Force", false).
		Execute().
		Get("Job", &job).
		Get("ReturnValue", &res).End()
	if err != nil {
		return err
	}
	return waitVMResult(res, srv, job)
}

func (vm *VirtualMachine) Start() error {
	var (
		srv *wmi.Service
		err error
		job *wmi.Instance
		res int32
	)

	if s := vm.EnabledState; !Disabled.equal(s) {
		if Enabled.equal(s) {
			return ErrMachineAlreadyRunning
		} else if Starting.equal(s) {
			return ErrMachineAlreadyRunning
		}
		return errors.New("machine not in a state to start")
	}

	if srv, err = getService(srv); err != nil {
		return err
	}
	defer srv.Close()

	instance, err := srv.GetObject(vm.Path())
	if err != nil {
		return err
	}
	defer instance.Close()

	// https://learn.microsoft.com/en-us/windows/win32/hyperv_v2/cim-concretejob-requeststatechange
	if err := wmiext.BeginInvoke(srv, instance, "RequestStateChange").
		Set("RequestedState", uint16(start)).
		Set("TimeoutPeriod", &time.Time{}).
		Execute().
		Get("Job", &job).
		Get("ReturnValue", &res).End(); err != nil {
		return err
	}
	return waitVMResult(res, srv, job)
}

func getService(_ *wmi.Service) (*wmi.Service, error) {
	// any reason why when we instantiate a vm, we should NOT just embed a service?
	return wmi.NewLocalService(HyperVNamespace)
}

func (vm *VirtualMachine) list() ([]*HyperVConfig, error) {

	return nil, ErrNotImplemented
}

func (vm *VirtualMachine) GetConfig() (*HyperVConfig, error) {
	var (
		err error
		//job *wmi.Instance
		res int32
		srv *wmi.Service
	)

	//summary := SummaryInformation{}

	if srv, err = wmi.NewLocalService(HyperVNamespace); err != nil {
		return nil, err
	}
	wmiInst, err := wmiext.FindFirstRelatedInstance(srv, vm.Path(), "Msvm_VirtualSystemSettingData")
	if err != nil {
		return nil, err
	}
	defer wmiInst.Close()

	path, err := wmiext.ConvertToPath(wmiInst)
	if err != nil {
		return nil, err
	}

	imms, err := wmiext.GetSingletonInstance(srv, "Msvm_VirtualSystemManagementService")
	if err != nil {
		return nil, err
	}
	defer imms.Close()
	var foo []string
	err = wmiext.BeginInvoke(srv, imms, "GetSummaryInformation").
		Set("RequestedInformation", []uint32{103}).
		Set("SettingData", []string{path}).
		Execute().
		Get("ReturnValue", &res).
		Get("SummaryInformation", &foo).End()
	if err != nil {
		fmt.Println(res)
		return nil, err
	}
	//names, err := wmiInst.
	//if err != nil {
	//	return nil, err
	//}
	//
	//fmt.Println(names)
	//
	//state := EnabledState(vm.EnabledState)

	config := HyperVConfig{
		//CPUs:     0,
		//Created:  time.Time{},
		//DiskSize: 0,
		//LastUp:   time.Time{},
		//Memory:   0,
		//Running:  state == Enabled,
		//Starting: state == Quiesce,
		//State:    state,
	}
	return &config, nil
}

// NewVirtualMachine creates a new vm in hyperv
// decided to not return a *VirtualMachine here because of how Podman is
// likely to use this.  this could be easily added if desirable
func (vmm *VirtualMachineManager) NewVirtualMachine(name string, config *HardwareConfig) error {
	exists, err := vmm.Exists(name)
	if err != nil {
		return err
	}
	if exists {
		return ErrMachineAlreadyExists
	}

	// TODO I gotta believe there are naming restrictions for vms in hyperv?
	// TODO If something fails during creation, do we rip things down or follow precedent from other machines?  user deletes things

	systemSettings, err := NewSystemSettingsBuilder().
		PrepareSystemSettings(name, nil).
		PrepareMemorySettings(func(ms *MemorySettings) {
			//ms.DynamicMemoryEnabled = false
			//ms.VirtualQuantity = 8192 // Startup memory
			//ms.Reservation = config.Memory // min

			// The API seems to require both of these even
			// when not using dynamic memory
			ms.Limit = config.Memory
			ms.VirtualQuantity = config.Memory
		}).
		PrepareProcessorSettings(func(ps *ProcessorSettings) {
			ps.VirtualQuantity = config.CPUs // 4 cores
		}).
		Build()
	if err != nil {
		return err
	}

	//if err := vmm.CreateVhdxFile(config.DiskPath, config.DiskSize*1024*1024*1024); err != nil {
	//	return err
	//}
	if err := NewDriveSettingsBuilder(systemSettings).
		AddScsiController().
		AddSyntheticDiskDrive(0).
		DefineVirtualHardDisk(config.DiskPath, func(vhdss *VirtualHardDiskStorageSettings) {
			// set extra params like
			// vhdss.IOPSLimit = 5000
		}).
		Finish(). // disk
		Finish(). // drive
		//AddSyntheticDvdDrive(1).
		//DefineVirtualDvdDisk(isoFile).
		//Finish(). // disk
		//Finish(). // drive
		Finish(). // controller
		Complete(); err != nil {
		return err
	}
	// Add default network connection
	if err := NewNetworkSettingsBuilder(systemSettings).
		AddSyntheticEthernetPort(nil).
		AddEthernetPortAllocation(""). // "" = connect to default switch
		Finish().                      // allocation
		Finish().                      // port
		Complete(); err != nil {
		return err
	}
	return nil
}

func (vm *VirtualMachine) remove() (int32, error) {
	var (
		err error
		res int32
		srv *wmi.Service
	)

	// Check for disabled/stopped state
	if !Disabled.equal(vm.EnabledState) {
		return -1, ErrMachineStateInvalid
	}
	if srv, err = wmi.NewLocalService(HyperVNamespace); err != nil {
		return -1, err
	}

	wmiInst, err := wmiext.FindFirstRelatedInstance(srv, vm.Path(), "Msvm_VirtualSystemSettingData")
	if err != nil {
		return -1, err
	}
	defer wmiInst.Close()

	path, err := wmiext.ConvertToPath(wmiInst)
	if err != nil {
		return -1, err
	}

	vsms, err := wmiext.GetSingletonInstance(srv, "Msvm_VirtualSystemManagementService")
	if err != nil {
		return -1, err
	}
	defer wmiInst.Close()

	var (
		job             *wmi.Instance
		resultingSystem string
	)
	// https://learn.microsoft.com/en-us/windows/win32/hyperv_v2/cim-virtualsystemmanagementservice-destroysystem
	if err := wmiext.BeginInvoke(srv, vsms, "DestroySystem").
		Set("AffectedSystem", path).
		Execute().
		Get("Job", &job).
		Get("ResultingSystem", &resultingSystem).
		Get("ReturnValue", &res).End(); err != nil {
		return -1, err
	}

	// do i have this correct? you can get an error without a result?
	if err := waitVMResult(res, srv, job); err != nil {
		return -1, err
	}
	return res, nil
}

func (vm *VirtualMachine) Remove(diskPath string) error {
	res, err := vm.remove()
	if err != nil {
		return err
	}
	if DestroySystemResult(res) == VMDestroyCompletedwithNoError {
		// Remove disk only if we were given one
		if len(diskPath) > 0 {
			if err := os.Remove(diskPath); err != nil {
				return err
			}
		}
		return nil
	}
	return fmt.Errorf("failed to destroy system %s: %s", vm.Name, DestroySystemResult(res).Reason())

}

func (vm *VirtualMachine) State() EnabledState {
	return EnabledState(vm.EnabledState)
}

func (vm *VirtualMachine) IsStarting() bool {
	return Starting.equal(vm.EnabledState)
}
