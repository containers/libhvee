// +build windows
package wmi

// Some predefined classes for WMI objects on Windows.  These are taken from the Microsoft documentation, where:
//  - Variable types are renamed:
//     - String -> string
//     - Boolean -> bool
//     - boolean -> bool
//     - sint* -> int*
//     - datetime -> string (REVISIT: can we use Go's type?)
//  - The array operator, [], is moved from the parameter name to its type
//  - Default values following an "=" are removed
//  - And finally, the parameter name and type are swapped

// Win32_BIOS WMI class
type Win32_BIOS struct {
	BiosCharacteristics            []uint16
	BIOSVersion                    []string
	BuildNumber                    string
	Caption                        string
	CodeSet                        string
	CurrentLanguage                string
	Description                    string
	EmbeddedControllerMajorVersion uint8
	EmbeddedControllerMinorVersion uint8
	IdentificationCode             string
	InstallableLanguages           uint16
	InstallDate                    string
	LanguageEdition                string
	ListOfLanguages                []string
	Manufacturer                   string
	Name                           string
	OtherTargetOS                  string
	PrimaryBIOS                    bool
	ReleaseDate                    string
	SerialNumber                   string
	SMBIOSBIOSVersion              string
	SMBIOSMajorVersion             uint16
	SMBIOSMinorVersion             uint16
	SMBIOSPresent                  bool
	SoftwareElementID              string
	SoftwareElementState           uint16
	Status                         string
	SystemBiosMajorVersion         uint8
	SystemBiosMinorVersion         uint8
	TargetOperatingSystem          uint16
	Version                        string
}

// Win32_DiskDrive WMI class
type Win32_DiskDrive struct {
	Availability                uint16
	BytesPerSector              uint32
	Capabilities                []uint16
	CapabilityDescriptions      []string
	Caption                     string
	CompressionMethod           string
	ConfigManagerErrorCode      uint32
	ConfigManagerUserConfig     bool
	CreationClassName           string
	DefaultBlockSize            uint64
	Description                 string
	DeviceID                    string
	ErrorCleared                bool
	ErrorDescription            string
	ErrorMethodology            string
	FirmwareRevision            string
	Index                       uint32
	InstallDate                 string
	InterfaceType               string
	LastErrorCode               uint32
	Manufacturer                string
	MaxBlockSize                uint64
	MaxMediaSize                uint64
	MediaLoaded                 bool
	MediaType                   string
	MinBlockSize                uint64
	Model                       string
	Name                        string
	NeedsCleaning               bool
	NumberOfMediaSupported      uint32
	Partitions                  uint32
	PNPDeviceID                 string
	PowerManagementCapabilities []uint16
	PowerManagementSupported    bool
	SCSIBus                     uint32
	SCSILogicalUnit             uint16
	SCSIPort                    uint16
	SCSITargetId                uint16
	SectorsPerTrack             uint32
	SerialNumber                string
	Signature                   uint32
	Size                        uint64
	Status                      string
	StatusInfo                  uint16
	SystemCreationClassName     string
	SystemName                  string
	TotalCylinders              uint64
	TotalHeads                  uint32
	TotalSectors                uint64
	TotalTracks                 uint64
	TracksPerCylinder           uint32
}

// Win32_NetworkAdapter WMI class
type Win32_NetworkAdapter struct {
	AdapterType                 string
	AdapterTypeID               uint16
	AutoSense                   bool
	Availability                uint16
	Caption                     string
	ConfigManagerErrorCode      uint32
	ConfigManagerUserConfig     bool
	CreationClassName           string
	Description                 string
	DeviceID                    string
	ErrorCleared                bool
	ErrorDescription            string
	GUID                        string
	Index                       uint32
	InstallDate                 string
	Installed                   bool
	InterfaceIndex              uint32
	LastErrorCode               uint32
	MACAddress                  string
	Manufacturer                string
	MaxNumberControlled         uint32
	MaxSpeed                    uint64
	Name                        string
	NetConnectionID             string
	NetConnectionStatus         uint16
	NetEnabled                  bool
	NetworkAddresses            []string
	PermanentAddress            string
	PhysicalAdapter             bool
	PNPDeviceID                 string
	PowerManagementCapabilities []uint16
	PowerManagementSupported    bool
	ProductName                 string
	ServiceName                 string
	Speed                       uint64
	Status                      string
	StatusInfo                  uint16
	SystemCreationClassName     string
	SystemName                  string
	TimeOfLastReset             string
}

// Win32_Processor WMI class
type Win32_Processor struct {
	AddressWidth                            uint16
	Architecture                            uint16
	AssetTag                                string
	Availability                            uint16
	Caption                                 string
	Characteristics                         uint32
	ConfigManagerErrorCode                  uint32
	ConfigManagerUserConfig                 bool
	CpuStatus                               uint16
	CreationClassName                       string
	CurrentClockSpeed                       uint32
	CurrentVoltage                          uint16
	DataWidth                               uint16
	Description                             string
	DeviceID                                string
	ErrorCleared                            bool
	ErrorDescription                        string
	ExtClock                                uint32
	Family                                  uint16
	InstallDate                             string
	L2CacheSize                             uint32
	L2CacheSpeed                            uint32
	L3CacheSize                             uint32
	L3CacheSpeed                            uint32
	LastErrorCode                           uint32
	Level                                   uint16
	LoadPercentage                          uint16
	Manufacturer                            string
	MaxClockSpeed                           uint32
	Name                                    string
	NumberOfCores                           uint32
	NumberOfEnabledCore                     uint32
	NumberOfLogicalProcessors               uint32
	OtherFamilyDescription                  string
	PartNumber                              string
	PNPDeviceID                             string
	PowerManagementCapabilities             []uint16
	PowerManagementSupported                bool
	ProcessorId                             string
	ProcessorType                           uint16
	Revision                                uint16
	Role                                    string
	SecondLevelAddressTranslationExtensions bool
	SerialNumber                            string
	SocketDesignation                       string
	Status                                  string
	StatusInfo                              uint16
	Stepping                                string
	SystemCreationClassName                 string
	SystemName                              string
	ThreadCount                             uint32
	UniqueId                                string
	UpgradeMethod                           uint16
	Version                                 string
	VirtualizationFirmwareEnabled           bool
	VMMonitorModeExtensions                 bool
	VoltageCaps                             uint32
}

// Win32_ComputerSystem WMI class
type Win32_ComputerSystem struct {
	AdminPasswordStatus         uint16
	AutomaticManagedPagefile    bool
	AutomaticResetBootOption    bool
	AutomaticResetCapability    bool
	BootOptionOnLimit           uint16
	BootOptionOnWatchDog        uint16
	BootROMSupported            bool
	BootupState                 string
	BootStatus                  []uint16
	Caption                     string
	ChassisBootupState          uint16
	ChassisSKUNumber            string
	CreationClassName           string
	CurrentTimeZone             int16
	DaylightInEffect            bool
	Description                 string
	DNSHostName                 string
	Domain                      string
	DomainRole                  uint16
	EnableDaylightSavingsTime   bool
	FrontPanelResetStatus       uint16
	HypervisorPresent           bool
	InfraredSupported           bool
	InitialLoadInfo             []string
	InstallDate                 string
	KeyboardPasswordStatus      uint16
	LastLoadInfo                string
	Manufacturer                string
	Model                       string
	Name                        string
	NameFormat                  string
	NetworkServerModeEnabled    bool
	NumberOfLogicalProcessors   uint32
	NumberOfProcessors          uint32
	OEMLogoBitmap               []uint8
	OEMStringArray              []string
	PartOfDomain                bool
	PauseAfterReset             int64
	PCSystemType                uint16
	PCSystemTypeEx              uint16
	PowerManagementCapabilities []uint16
	PowerManagementSupported    bool
	PowerOnPasswordStatus       uint16
	PowerState                  uint16
	PowerSupplyState            uint16
	PrimaryOwnerContact         string
	PrimaryOwnerName            string
	ResetCapability             uint16
	ResetCount                  int16
	ResetLimit                  int16
	Roles                       []string
	Status                      string
	SupportContactDescription   []string
	SystemFamily                string
	SystemSKUNumber             string
	SystemStartupDelay          uint16
	SystemStartupOptions        []string
	SystemStartupSetting        uint8
	SystemType                  string
	ThermalState                uint16
	TotalPhysicalMemory         uint64
	UserName                    string
	WakeUpType                  uint16
	Workgroup                   string
}

// Win32_ComputerSystemProduct WMI class
type Win32_ComputerSystemProduct struct {
	Caption           string
	Description       string
	IdentifyingNumber string
	Name              string
	SKUNumber         string
	Vendor            string
	Version           string
	UUID              string
}

// Win32_Directory WMI class
type Win32_Directory struct {
	Caption               string
	Description           string
	InstallDate           string
	Name                  string
	Status                string
	AccessMask            uint32
	Archive               bool
	Compressed            bool
	CompressionMethod     string
	CreationClassName     string
	CreationDate          string
	CSCreationClassName   string
	CSName                string
	Drive                 string
	EightDotThreeFileName string
	Encrypted             bool
	EncryptionMethod      string
	Extension             string
	FileName              string
	FileSize              uint64
	FileType              string
	FSCreationClassName   string
	FSName                string
	Hidden                bool
	InUseCount            uint64
	LastAccessed          string
	LastModified          string
	Path                  string
	Readable              bool
	System                bool
	Writeable             bool
}

// Win32_OperatingSystem WMI class
type Win32_OperatingSystem struct {
	BootDevice                                string
	BuildNumber                               string
	BuildType                                 string
	Caption                                   string
	CodeSet                                   string
	CountryCode                               string
	CreationClassName                         string
	CSCreationClassName                       string
	CSDVersion                                string
	CSName                                    string
	CurrentTimeZone                           int16
	DataExecutionPrevention_Available         bool
	DataExecutionPrevention_32BitApplications bool
	DataExecutionPrevention_Drivers           bool
	DataExecutionPrevention_SupportPolicy     uint8
	Debug                                     bool
	Description                               string
	Distributed                               bool
	EncryptionLevel                           uint32
	ForegroundApplicationBoost                uint8
	FreePhysicalMemory                        uint64
	FreeSpaceInPagingFiles                    uint64
	FreeVirtualMemory                         uint64
	InstallDate                               string
	LargeSystemCache                          uint32
	LastBootUpTime                            string
	LocalDateTime                             string
	Locale                                    string
	Manufacturer                              string
	MaxNumberOfProcesses                      uint32
	MaxProcessMemorySize                      uint64
	MUILanguages                              []string
	Name                                      string
	NumberOfLicensedUsers                     uint32
	NumberOfProcesses                         uint32
	NumberOfUsers                             uint32
	OperatingSystemSKU                        uint32
	Organization                              string
	OSArchitecture                            string
	OSLanguage                                uint32
	OSProductSuite                            uint32
	OSType                                    uint16
	OtherTypeDescription                      string
	PAEEnabled                                bool
	PlusProductID                             string
	PlusVersionNumber                         string
	PortableOperatingSystem                   bool
	Primary                                   bool
	ProductType                               uint32
	RegisteredUser                            string
	SerialNumber                              string
	ServicePackMajorVersion                   uint16
	ServicePackMinorVersion                   uint16
	SizeStoredInPagingFiles                   uint64
	Status                                    string
	SuiteMask                                 uint32
	SystemDevice                              string
	SystemDirectory                           string
	SystemDrive                               string
	TotalSwapSpaceSize                        uint64
	TotalVirtualMemorySize                    uint64
	TotalVisibleMemorySize                    uint64
	Version                                   string
	WindowsDirectory                          string
	QuantumLength                             uint8
	QuantumType                               uint8
}

// Win32_Process WMI class
type Win32_Process struct {
	CreationClassName          string
	Caption                    string
	CommandLine                string
	CreationDate               string
	CSCreationClassName        string
	CSName                     string
	Description                string
	ExecutablePath             string
	ExecutionState             uint16
	Handle                     string
	HandleCount                uint32
	InstallDate                string
	KernelModeTime             uint64
	MaximumWorkingSetSize      uint32
	MinimumWorkingSetSize      uint32
	Name                       string
	OSCreationClassName        string
	OSName                     string
	OtherOperationCount        uint64
	OtherTransferCount         uint64
	PageFaults                 uint32
	PageFileUsage              uint32
	ParentProcessId            uint32
	PeakPageFileUsage          uint32
	PeakVirtualSize            uint64
	PeakWorkingSetSize         uint32
	Priority                   uint32
	PrivatePageCount           uint64
	ProcessId                  uint32
	QuotaNonPagedPoolUsage     uint32
	QuotaPagedPoolUsage        uint32
	QuotaPeakNonPagedPoolUsage uint32
	QuotaPeakPagedPoolUsage    uint32
	ReadOperationCount         uint64
	ReadTransferCount          uint64
	SessionId                  uint32
	Status                     string
	TerminationDate            string
	ThreadCount                uint32
	UserModeTime               uint64
	VirtualSize                uint64
	WindowsVersion             string
	WorkingSetSize             uint64
	WriteOperationCount        uint64
	WriteTransferCount         uint64
}

// Win32_Registry WMI class
type Win32_Registry struct {
	Caption      string
	Description  string
	InstallDate  string
	Status       string
	CurrentSize  uint32
	MaximumSize  uint32
	Name         string
	ProposedSize uint32
}

// Win32_Product WMI class
type Win32_Product struct {
	AssignmentType    uint16
	Caption           string
	Description       string
	IdentifyingNumber string
	InstallDate       string
	InstallDate2      string
	InstallLocation   string
	InstallState      int16
	HelpLink          string
	HelpTelephone     string
	InstallSource     string
	Language          string
	LocalPackage      string
	Name              string
	PackageCache      string
	PackageCode       string
	PackageName       string
	ProductID         string
	RegOwner          string
	RegCompany        string
	SKUNumber         string
	Transforms        string
	URLInfoAbout      string
	URLUpdateInfo     string
	Vendor            string
	WordCount         uint32
	Version           string
}

// CIM_DataFile WMI class
type CIM_DataFile struct {
	Caption               string
	Description           string
	InstallDate           string
	Status                string
	AccessMask            uint32
	Archive               bool
	Compressed            bool
	CompressionMethod     string
	CreationClassName     string
	CreationDate          string
	CSCreationClassName   string
	CSName                string
	Drive                 string
	EightDotThreeFileName string
	Encrypted             bool
	EncryptionMethod      string
	Name                  string
	Extension             string
	FileName              string
	FileSize              uint64
	FileType              string
	FSCreationClassName   string
	FSName                string
	Hidden                bool
	InUseCount            uint64
	LastAccessed          string
	LastModified          string
	Path                  string
	Readable              bool
	System                bool
	Writeable             bool
	Manufacturer          string
	Version               string
}
