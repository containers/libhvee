package hypervctl

import (
	"fmt"

	"github.com/containers/libhvee/pkg/wmiext"
	"github.com/drtimf/wmi"
)

type SystemSettingsBuilder struct {
	systemSettings    *SystemSettings
	processorSettings *ProcessorSettings
	memorySettings    *MemorySettings
	err               error
}

func NewSystemSettingsBuilder() *SystemSettingsBuilder {
	return &SystemSettingsBuilder{}
}

func (builder *SystemSettingsBuilder) PrepareSystemSettings(name string, beforeAdd func(*SystemSettings)) *SystemSettingsBuilder {
	if builder.err != nil {
		return builder
	}

	if builder.systemSettings == nil {
		settings := DefaultSystemSettings()
		settings.ElementName = name
		builder.systemSettings = settings
	}

	if beforeAdd != nil {
		beforeAdd(builder.systemSettings)
	}

	return builder
}

func (builder *SystemSettingsBuilder) PrepareProcessorSettings(beforeAdd func(*ProcessorSettings)) *SystemSettingsBuilder {
	if builder.err != nil {
		return builder
	}

	if builder.processorSettings == nil {
		settings, err := fetchDefaultProcessorSettings()
		if err != nil {
			builder.err = err
			return builder
		}
		builder.processorSettings = settings
	}

	if beforeAdd != nil {
		beforeAdd(builder.processorSettings)
	}

	return builder
}

func (builder *SystemSettingsBuilder) PrepareMemorySettings(beforeAdd func(*MemorySettings)) *SystemSettingsBuilder {
	if builder.err != nil {
		return builder
	}

	if builder.memorySettings == nil {
		settings, err := fetchDefaultMemorySettings()
		if err != nil {
			builder.err = err
			return builder
		}
		builder.memorySettings = settings
	}

	if beforeAdd != nil {
		beforeAdd(builder.memorySettings)
	}

	return builder
}

func (builder *SystemSettingsBuilder) Build() (*SystemSettings, error) {
	var service *wmi.Service
	var err error

	if builder.PrepareSystemSettings("unnamed-vm", nil).
		PrepareProcessorSettings(nil).
		PrepareMemorySettings(nil).err != nil {
		return nil, err
	}

	if service, err = wmi.NewLocalService(HyperVNamespace); err != nil {
		return nil, err
	}
	defer service.Close()

	systemSettingsInst, err := wmiext.SpawnInstance(service, "Msvm_VirtualSystemSettingData")
	if err != nil {
		return nil, err
	}
	defer systemSettingsInst.Close()

	err = wmiext.InstancePutAll(systemSettingsInst, builder.systemSettings)
	if err != nil {
		return nil, err
	}

	memoryStr, err := createMemorySettings(builder.memorySettings)
	if err != nil {
		return nil, err
	}

	processorStr, err := createProcessorSettings(builder.processorSettings)
	if err != nil {
		return nil, err
	}

	vsms, err := wmiext.GetSingletonInstance(service, VirtualSystemManagementService)
	if err != nil {
		return nil, err
	}
	defer vsms.Close()

	systemStr := wmiext.GetCimText(systemSettingsInst)

	var job *wmi.Instance
	var res int32
	var resultingSystem string
	err = wmiext.BeginInvoke(service, vsms, "DefineSystem").
		Set("SystemSettings", systemStr).
		Set("ResourceSettings", []string{memoryStr, processorStr}).
		Execute().
		Get("Job", &job).
		Get("ResultingSystem", &resultingSystem).
		Get("ReturnValue", &res).End()

	if err != nil {
		return nil, fmt.Errorf("Failed to define system: %w", err)
	}

	err = waitVMResult(res, service, job)
	if err != nil {
		return nil, fmt.Errorf("Failed to define system: %w", err)
	}

	newSettings, err := wmiext.FindFirstRelatedInstance(service, resultingSystem, "Msvm_VirtualSystemSettingData")
	if err != nil {
		return nil, err
	}
	path, err := wmiext.ConvertToPath(newSettings)
	if err != nil {
		return nil, err
	}

	if err = wmiext.GetObjectAsObject(service, path, builder.systemSettings); err != nil {
		return nil, err
	}

	return builder.systemSettings, nil
}
