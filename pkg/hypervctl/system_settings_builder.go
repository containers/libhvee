package hypervctl

import (
	"errors"
	"fmt"

	"github.com/drtimf/wmi"
	"github.com/n1hility/hypervctl/pkg/wmiext"
)

type SystemSettingsBuilder struct {
	systemSettings    *SystemSettings
	processorSettings *ProcessorSettings
	memorySettings    *MemorySettings
}

func NewSystemSettingsBuilder() *SystemSettingsBuilder {
	return &SystemSettingsBuilder{}
}

func (builder *SystemSettingsBuilder) PrepareSystemSettings(name string) *SystemSettings {
	settings := DefaultSystemSettings()
	settings.ElementName = name
	builder.systemSettings = settings
	return settings
}

func (builder *SystemSettingsBuilder) PrepareProcessorSettings() (*ProcessorSettings, error) {
	var err error
	var settings *ProcessorSettings

	if builder.processorSettings != nil {
		return builder.processorSettings, nil
	}

	settings, err = fetchDefaultProcessorSettings()
	if err == nil {
		builder.processorSettings = settings
	}

	return settings, err
}

func (builder *SystemSettingsBuilder) PrepareMemorySettings() (*MemorySettings, error) {
	var err error
	var settings *MemorySettings

	if builder.processorSettings != nil {
		return builder.memorySettings, nil
	}

	settings, err = fetchDefaultMemorySettings()
	if err == nil {
		builder.memorySettings = settings
	}

	return settings, err
}

func (builder *SystemSettingsBuilder) Build() (*SystemSettings, error) {
	var service *wmi.Service
	var err error

	if service, err = wmi.NewLocalService(HyperVNamespace); err != nil {
		return nil, err
	}
	defer service.Close()

	systemSettings := builder.systemSettings
	if systemSettings == nil {
		return nil, errors.New("prepareSettings not called on builder")
	}

	processorSettings := builder.processorSettings
	if processorSettings == nil {
		return nil, errors.New("preparProcessorSettings not called on builder")
	}

	memorySettings := builder.memorySettings
	if memorySettings == nil {
		return nil, errors.New("prepareMemorySettings not called on builder")
	}

	systemSettingsInst, err := wmiext.SpawnInstance(service, "Msvm_VirtualSystemSettingData")
	if err != nil {
		return nil, err
	}
	defer systemSettingsInst.Close()

	err = wmiext.InstancePutAll(systemSettingsInst, systemSettings)
	if err != nil {
		return nil, err
	}

	memoryStr, err := createMemorySettings(memorySettings)
	if err != nil {
		return nil, err
	}

	processorStr, err := createProcessorSettings(processorSettings)
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

	if err = wmiext.GetObjectAsObject(service, path, systemSettings); err != nil {
		return nil, err
	}

	return systemSettings, nil
}
