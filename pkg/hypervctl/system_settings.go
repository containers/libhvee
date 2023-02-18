package hypervctl

import (
	"fmt"
	"time"

	"github.com/drtimf/wmi"
	"github.com/n1hility/hypervctl/pkg/wmiext"
)

type SystemSettings struct {
	S__PATH                              string
	InstanceID                           string
	Caption                              string // = "Virtual Machine Settings"
	Description                          string
	ElementName                          string
	VirtualSystemIdentifier              string
	VirtualSystemType                    string
	Notes                                []string
	CreationTime                         time.Time
	ConfigurationID                      string
	ConfigurationDataRoot                string
	ConfigurationFile                    string
	SnapshotDataRoot                     string
	SuspendDataRoot                      string
	SwapFileDataRoot                     string
	LogDataRoot                          string
	AutomaticStartupAction               uint16 // non-zero
	AutomaticStartupActionDelay          time.Time
	AutomaticStartupActionSequenceNumber uint16
	AutomaticShutdownAction              uint16 // non-zero
	AutomaticRecoveryAction              uint16 // non-zero
	RecoveryFile                         string
	BIOSGUID                             string
	BIOSSerialNumber                     string
	BaseBoardSerialNumber                string
	ChassisSerialNumber                  string
	Architecture                         string
	ChassisAssetTag                      string
	BIOSNumLock                          bool
	BootOrder                            []uint16
	Parent                               string
	UserSnapshotType                     uint16 // non-zero
	IsSaved                              bool
	AdditionalRecoveryInformation        string
	AllowFullSCSICommandSet              bool
	DebugChannelId                       uint32
	DebugPortEnabled                     uint16
	DebugPort                            uint32
	Version                              string
	IncrementalBackupEnabled             bool
	VirtualNumaEnabled                   bool
	AllowReducedFcRedundancy             bool // = False
	VirtualSystemSubType                 string
	BootSourceOrder                      []string
	PauseAfterBootFailure                bool
	NetworkBootPreferredProtocol         uint16 // non-zero
	GuestControlledCacheTypes            bool
	AutomaticSnapshotsEnabled            bool
	IsAutomaticSnapshot                  bool
	GuestStateFile                       string
	GuestStateDataRoot                   string
	LockOnDisconnect                     bool
	ParentPackage                        string
	AutomaticCriticalErrorActionTimeout  time.Time
	AutomaticCriticalErrorAction         uint16
	ConsoleMode                          uint16
	SecureBootEnabled                    bool
	SecureBootTemplateId                 string
	LowMmioGapSize                       uint64
	HighMmioGapSize                      uint64
	EnhancedSessionTransportType         uint16
}

func DefaultSystemSettings() *SystemSettings {
	return &SystemSettings{
		// setup all non-zero settings
		AutomaticStartupAction:       2,    // no auto-start
		AutomaticShutdownAction:      4,    // shutdown
		AutomaticRecoveryAction:      3,    // restart
		UserSnapshotType:             2,    // no snapshotting
		NetworkBootPreferredProtocol: 4096, // ipv4 for pxe
		VirtualSystemSubType:         "Microsoft:Hyper-V:SubType:2",
	}

}

func (s *SystemSettings) Path() string {
	return s.S__PATH
}

func (systemSettings *SystemSettings) AddScsiController() (*ScsiControllerSettings, error) {
	const scsiControllerType = "Microsoft:Hyper-V:Synthetic SCSI Controller"
	controller := &ScsiControllerSettings{}

	if err := systemSettings.createSystemResourceInternal(controller, scsiControllerType, nil); err != nil {
		return nil, err
	}

	controller.systemSettings = systemSettings
	return controller, nil
}

func (systemSettings *SystemSettings) createSystemResourceInternal(settings interface{}, resourceType string, cb func()) error {
	var service *wmi.Service
	var err error
	if service, err = wmi.NewLocalService(HyperVNamespace); err != nil {
		return err
	}
	defer service.Close()

	if err = populateDefaults(resourceType, settings); err != nil {
		return err
	}

	if cb != nil {
		cb()
	}

	resourceStr, err := createResourceSettingGeneric(settings, resourceType)
	if err != nil {
		return err
	}	

	path, err := addResource(service, systemSettings.Path(), resourceStr)
	if err != nil {
		return err
	}

	err = wmiext.GetObjectAsObject(service, path, settings)
	return err
}

func (systemSettings *SystemSettings) AddSyntheticEthernetPort(beforeAdd func(*SyntheticEthernetPortSettings)) (*SyntheticEthernetPortSettings, error) {
	const networkAdapterType = SyntheticEthernetPortResourceType
	port := &SyntheticEthernetPortSettings{}

	var cb func()
	if beforeAdd != nil {
		cb = func() {
			beforeAdd(port)
		}
	}

	if err := systemSettings.createSystemResourceInternal(port, networkAdapterType, cb); err != nil {
		return nil, err
	}

	port.systemSettings = systemSettings
	return port, nil
}

func addResource(service *wmi.Service, systemSettingPath string, resourceSettings string) (string, error) {
	vsms, err := wmiext.GetSingletonInstance(service, VirtualSystemManagementService)
	if err != nil {
		return "", err
	}
	defer vsms.Close()

	var res int32
	var resultingSettings []string
	var job *wmi.Instance
	err = wmiext.BeginInvoke(service, vsms, "AddResourceSettings").
		Set("AffectedConfiguration", systemSettingPath).
		Set("ResourceSettings", []string{resourceSettings}).
		Execute().
		Get("Job", &job).
		Get("ResultingResourceSettings", &resultingSettings).
		Get("ReturnValue", &res).End()

	if err != nil {
		return "", fmt.Errorf("AddResourceSettings failed: %w", err)
	}

	err = waitVMResult(res, service, job)

	if len(resultingSettings) > 0 {
		return resultingSettings[0], err
	}

	return "", err
}

func (s *SystemSettings) GetVM() (*VirtualMachine, error) {
	var service *wmi.Service
	var err error
	if service, err = wmi.NewLocalService(HyperVNamespace); err != nil {
		return nil, err
	}
	defer service.Close()

	inst, err := wmiext.FindFirstRelatedInstance(service, s.Path(), "Msvm_ComputerSystem")
	if err != nil {
		return nil, err
	}
	defer inst.Close()

	vm := &VirtualMachine{}

	if err = wmiext.InstanceGetAll(inst, vm); err != nil {
		return nil, err
	}

	return vm, nil
}
