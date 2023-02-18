package hypervctl

import (
	"fmt"

	"github.com/drtimf/wmi"
	"github.com/n1hility/hypervctl/pkg/wmiext"
)

const (
	HyperVNamespace                = "root\\virtualization\\v2"
	VirtualSystemManagementService = "Msvm_VirtualSystemManagementService"
)

type VirtualMachineManager struct {
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

func (*VirtualMachineManager) CreateVhdxFile(path string, maxSize uint64) error {
	var service *wmi.Service
	var err error
	if service, err = wmi.NewLocalService(HyperVNamespace); err != nil {
		return err
	}
	defer service.Close()

	settings := &VirtualHardDiskSettings{}
	settings.Format = 3
	settings.MaxInternalSize = maxSize
	settings.Type = 3
	settings.Path = path

	instance, err := wmiext.CreateInstance(service, "Msvm_VirtualHardDiskSettingData", settings)
	if err != nil {
		return err
	}
	defer instance.Close()
	settingsStr := wmiext.GetCimText(instance)

	imms, err := wmiext.GetSingletonInstance(service, "Msvm_ImageManagementService")
	if err != nil {
		return err
	}
	defer imms.Close()

	var job *wmi.Instance
	var ret int32
	err = wmiext.BeginInvoke(service, imms, "CreateVirtualHardDisk").
		Set("VirtualDiskSettingData", settingsStr).
		Execute().
		Get("Job", &job).
		Get("ReturnValue", &ret).
		End()

	if err != nil {
		return fmt.Errorf("Failed to create vhdx: %w", err)
	}

	return waitVMResult(ret, service, job)
}

