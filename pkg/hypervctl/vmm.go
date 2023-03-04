//go:build windows
// +build windows

package hypervctl

import (
	"fmt"
	"time"

	"github.com/containers/libhvee/pkg/wmiext"
)

const (
	HyperVNamespace                = "root\\virtualization\\v2"
	VirtualSystemManagementService = "Msvm_VirtualSystemManagementService"
)

// https://learn.microsoft.com/en-us/windows/win32/hyperv_v2/msvm-computersystem

type VirtualMachineManager struct {
}

func NewVirtualMachineManager() *VirtualMachineManager {
	return &VirtualMachineManager{}
}

func (*VirtualMachineManager) GetAll() ([]*VirtualMachine, error) {
	const wql = "Select * From Msvm_ComputerSystem Where Caption = 'Virtual Machine'"

	var service *wmiext.Service
	var err error
	if service, err = wmiext.NewLocalService(HyperVNamespace); err != nil {
		return []*VirtualMachine{}, err
	}
	defer service.Close()

	var enum *wmiext.Enum
	if enum, err = service.ExecQuery(wql); err != nil {
		return nil, err
	}
	defer enum.Close()

	var vms []*VirtualMachine
	for {
		vm := &VirtualMachine{}
		done, err := wmiext.NextObject(enum, vm)
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
func (vmm *VirtualMachineManager) Exists(name string) (bool, error) {
	vms, err := vmm.GetAll()
	if err != nil {
		return false, err
	}
	for _, i := range vms {
		// TODO should case be honored or ignored?
		if i.Name == name {
			return true, nil
		}
	}
	return false, nil
}

func (*VirtualMachineManager) GetMachine(name string) (*VirtualMachine, error) {
	const wql = "Select * From Msvm_ComputerSystem Where Caption = 'Virtual Machine' And ElementName='%s'"

	vm := &VirtualMachine{}
	var service *wmiext.Service
	var err error

	if service, err = wmiext.NewLocalService(HyperVNamespace); err != nil {
		return vm, err
	}
	defer service.Close()

	var enum *wmiext.Enum
	if enum, err = service.ExecQuery(fmt.Sprintf(wql, name)); err != nil {
		return nil, err
	}
	defer enum.Close()

	done, err := wmiext.NextObject(enum, vm)
	if err != nil {
		return vm, err
	}

	if done {
		return vm, fmt.Errorf("could not find virtual machine %q", name)
	}

	return vm, nil
}

func (*VirtualMachineManager) CreateVhdxFile(path string, maxSize uint64) error {
	var service *wmiext.Service
	var err error
	if service, err = wmiext.NewLocalService(HyperVNamespace); err != nil {
		return err
	}
	defer service.Close()

	settings := &VirtualHardDiskSettings{}
	settings.Format = 3
	settings.MaxInternalSize = maxSize
	settings.Type = 3
	settings.Path = path

	instance, err := service.CreateInstance("Msvm_VirtualHardDiskSettingData", settings)
	if err != nil {
		return err
	}
	defer instance.Close()
	settingsStr := instance.GetCimText()

	imms, err := service.GetSingletonInstance("Msvm_ImageManagementService")
	if err != nil {
		return err
	}
	defer imms.Close()

	var job *wmiext.Instance
	var ret int32
	err = imms.BeginInvoke("CreateVirtualHardDisk").
		In("VirtualDiskSettingData", settingsStr).
		Execute().
		Out("Job", &job).
		Out("ReturnValue", &ret).
		End()

	if err != nil {
		return fmt.Errorf("failed to create vhdx: %w", err)
	}

	return waitVMResult(ret, service, job)
}

func (*VirtualMachineManager) GetSummaryInformation() ([]SummaryInformation, error) {
	var service *wmiext.Service
	var err error
	if service, err = wmiext.NewLocalService(HyperVNamespace); err != nil {
		return nil, err
	}
	defer service.Close()

	vmms, err := service.GetSingletonInstance(VirtualSystemManagementService)
	if err != nil {
		return nil, err
	}
	defer vmms.Close()

	var summary []SummaryInformation

	err = vmms.BeginInvoke("GetSummaryInformation").
		In("RequestedInformation", []uint{0, 1, 2, 3, 4, 100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113}).Execute().
		Out("SummaryInformation", &summary).End()

	if err != nil {
		panic(err)
	}

	return summary, nil
}

type SummaryInformation struct {
	Name                  string
	ElementName           string
	EnabledState          uint16
	HealthState           uint16
	Notes                 string
	NumberOfProcessors    uint16
	ThumbnailImage        []uint8
	CreationTime          time.Time
	ProcessorLoad         uint16
	ProcessorLoadHistory  []uint16
	MemoryUsage           uint64
	MemoryAvailable       int32
	AvailableMemoryBuffer int32
	Heartbeat             uint16
	UpTime                uint64
	GuestOperatingSystem  string
	OperationalStatus     []uint16
	StatusDescriptions    []string
	//Snapshots []CIM_VirtualSystemSettingData
	//AsynchronousTasks []CIM_ConcreteJob
	AllocatedGPU string
}
