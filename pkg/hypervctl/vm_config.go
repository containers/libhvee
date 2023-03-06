package hypervctl

import (
	"time"
)

type vmState uint16

const (
	// Changes the state to 'Running'.
	start vmState = 2
	// Stops the job temporarily. The intention is to subsequently restart the job with 'Start'. It might be possible to
	// enter the 'Service' state while suspended. (This is job-specific.)
	suspend vmState = 3
	// Stops the job cleanly, saves data, preserves the state, and shuts down all underlying processes'
	// in an orderly manner.
	terminate vmState = 4
	//Terminates the job immediately with no requirement to save data or preserve the state.
	kill vmState = 5
)

type EnabledState uint16

const (
	// Unknown The state of the element could not be determined.
	Unknown EnabledState = 0
	// Other No description
	Other EnabledState = 1
	// Enabled The element is running.
	Enabled EnabledState = 2
	// Disabled The element is turned off.
	Disabled EnabledState = 3
	// ShuttingDown Shutting_Down The element is in the process of going to a Disabled state.
	ShuttingDown EnabledState = 4
	// NotApplicable The element does not support being enabled or disabled.
	NotApplicable EnabledState = 5
	// EnabledButOffline The element might be completing commands, and it will drop any new requests.
	EnabledButOffline EnabledState = 6
	// InTest The element is in a test state.
	InTest EnabledState = 7
	// Deferred The element might be completing commands, but it will queue any new requests.
	Deferred EnabledState = 8
	// Quiesce The element is enabled but in a restricted mode. The behavior of the element is similar to the Enabled state
	//(2), but it processes only a restricted set of commands. All other requests are queued.
	Quiesce EnabledState = 9
	// Starting The element is in the process of going to an Enabled state (2). New requests are queued.
	Starting EnabledState = 10
)

func (es EnabledState) String() string {
	switch es {
	case Unknown:
		return "unknown"
	case Other:
		return "other"
	case Enabled:
		return "enabled"
	case Disabled:
		return "disabled"
	case ShuttingDown:
		return "shutting down"
	case NotApplicable:
		return "not applicable"
	case EnabledButOffline:
		return "enabled but offline"
	case InTest:
		return "in test"
	case Deferred:
		return "deferred"
	case Quiesce:
		return "quiesce"
	}
	return "starting"
}

func (es EnabledState) equal(s uint16) bool {
	return EnabledState(s) == EnabledState(s)
}

// HyperVConfig describes physical attributes of the machine
type HyperVConfig struct {
	// Hardware configuration for HyperV
	Hardware HardwareConfig
	// state and descriptive info
	Status Statuses
}

type HardwareConfig struct {
	// CPUs to be assigned to the VM
	CPUs uint64
	// Diskpath is fully qualified location of the
	// bootable disk image
	DiskPath string
	// Disk size in gigabytes assigned to the vm
	DiskSize uint64
	// Memory in megabytes assigned to the vm
	Memory uint64
}

type Statuses struct {
	// time vm created
	Created time.Time
	// last time vm ran
	LastUp time.Time
	// is vm running
	Running bool
	// is vm starting up
	Starting bool
	// vm state
	State EnabledState
}

// SummaryInformation https://learn.microsoft.com/en-us/windows/win32/hyperv_v2/msvm-summaryinformation
type SummaryInformation struct {
	InstanceID                      string
	AllocatedGPU                    string
	Shielded                        bool
	AsynchronousTasks               []CIM_ConcreteJob
	CreationTime                    time.Time
	ElementName                     string
	EnabledState                    uint16
	OtherEnabledState               string
	GuestOperatingSystem            string
	HealthState                     uint16
	Heartbeat                       uint16
	MemoryUsage                     uint64
	MemoryAvailable                 int32 // changed from sint32
	AvailableMemoryBuffer           int32 // changed from sint32
	SwapFilesInUse                  bool
	Name                            string
	Notes                           string
	Version                         string
	NumberOfProcessors              uint16
	OperationalStatus               []uint16
	ProcessorLoad                   uint16
	ProcessorLoadHistory            []uint16
	Snapshots                       []SystemSettings
	StatusDescriptions              []string
	ThumbnailImage                  []uint8
	ThumbnailImageHeight            uint16
	ThumbnailImageWidth             uint16
	UpTime                          uint64
	ReplicationState                uint16
	ReplicationStateEx              []uint16
	ReplicationHealth               uint16
	ReplicationHealthEx             []uint16
	ReplicationMode                 uint16
	TestReplicaSystem               string // REF to CIM_ComputerSystem
	ApplicationHealth               uint16
	IntegrationServicesVersionState uint16
	MemorySpansPhysicalNumaNodes    bool
	ReplicationProviderId           []string
	EnhancedSessionModeState        uint16
	VirtualSwitchNames              []string
	VirtualSystemSubType            string
	HostComputerSystemName          string
}

// CIM_ConcreteJob https://learn.microsoft.com/en-us/windows/win32/hyperv_v2/msvm-concretejob
type CIM_ConcreteJob struct {
	InstanceID              string
	Caption                 string
	Description             string
	ElementName             string
	InstallDate             time.Time
	Name                    string
	OperationalStatus       []uint16 // = { 2 }
	StatusDescriptions      []string // = { "OK" }
	Status                  string
	HealthState             uint16 // = 5
	CommunicationStatus     uint16
	DetailedStatus          uint16
	OperatingStatus         uint16
	PrimaryStatus           uint16
	JobStatus               string
	TimeSubmitted           time.Time
	ScheduledStartTime      time.Time
	StartTime               time.Time
	ElapsedTime             time.Time
	JobRunTimes             uint32
	RunMonth                uint8
	RunDay                  int8 // changed from sint8
	RunDayOfWeek            int8 // changed from sint8
	RunStartInterval        time.Time
	LocalOrUtcTime          uint16
	UntilTime               time.Time
	Notify                  string
	Owner                   string
	Priority                uint32
	PercentComplete         uint16
	DeleteOnCompletion      bool
	ErrorCode               uint16
	ErrorDescription        string
	ErrorSummaryDescription string
	RecoveryAction          uint16
	OtherRecoveryAction     string
	JobState                uint16
	TimeOfLastStateChange   time.Time
	TimeBeforeRemoval       time.Time // =
	Cancellable             bool
	JobType                 uint16
}

// CIM_ComputerSystem https://learn.microsoft.com/en-us/windows/win32/cimwin32prov/cim-computersystem
type CIM_ComputerSystem struct {
	Caption             string
	Description         string
	InstallDate         time.Time
	Status              string
	CreationClassName   string
	Name                string
	PrimaryOwnerContact string
	PrimaryOwnerName    string
	Roles               []string
	NameFormat          string
}
