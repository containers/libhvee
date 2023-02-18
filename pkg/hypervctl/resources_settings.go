package hypervctl

import (
	"errors"
	"fmt"

	"github.com/drtimf/wmi"
	"github.com/n1hility/hypervctl/pkg/wmiext"
)

type ResourceSettings struct {
	S__PATH                  string
	InstanceID               string // = "Microsoft:GUID\DeviceSpecificData"
	Caption                  string
	Description              string
	ElementName              string
	ResourceType             uint16
	OtherResourceType        string
	ResourceSubType          string
	PoolID                   string
	ConsumerVisibility       uint16
	HostResource             []string
	AllocationUnits          string
	VirtualQuantity          uint64
	Reservation              uint64
	Limit                    uint64
	Weight                   uint32
	AutomaticAllocation      bool
	AutomaticDeallocation    bool
	Parent                   string
	Connection               []string
	Address                  string
	MappingBehavior          uint16
	AddressOnParent          string
	VirtualQuantityUnits     string   // = "count"
	VirtualSystemIdentifiers []string // = { "GUID" }
}

func (s *ResourceSettings) setParent(parent string) {
	s.Parent = parent
}

func (s *ResourceSettings) setAddressOnParent(address string) {
	s.AddressOnParent = address
}

func (s *ResourceSettings) Path() string {
	return s.S__PATH
}

func createResourceSettingGeneric(settings interface{}, resourceType string) (string, error) {
	var service *wmi.Service
	var err error
	if service, err = wmi.NewLocalService(HyperVNamespace); err != nil {
		return "", err
	}

	ref, err := findResourceDefaults(service, resourceType)
	if err != nil {
		return "", err
	}

	resource, err := service.GetObject(ref)
	if err != nil {
		return "", err
	}

	defer resource.Close()
	resource, err = wmiext.CloneInstance(resource)
	if err != nil {
		return "", err
	}
	defer resource.Close()

	if err = wmiext.InstancePutAll(resource, settings); err != nil {
		return "", err
	}

	return wmiext.GetCimText(resource), nil
}

func populateDefaults(subType string, settings interface{}) error {
	var service *wmi.Service
	var err error
	if service, err = wmi.NewLocalService(HyperVNamespace); err != nil {
		return err
	}
	defer service.Close()

	ref, err := findResourceDefaults(service, subType)
	if err != nil {
		return err
	}

	return wmiext.GetObjectAsObject(service, ref, settings)
}

func findResourceDefaults(service *wmi.Service, subType string) (string, error) {
	wql := fmt.Sprintf("SELECT * FROM Msvm_AllocationCapabilities WHERE ResourceSubType = '%s'", subType)
	instance, err := wmiext.FindFirstInstance(service, wql)
	if err != nil {
		return "", err
	}
	defer instance.Close()

	path, err := wmiext.ConvertToPath(instance)
	if err != nil {
		return "", err
	}

	enum, err := service.ExecQuery(fmt.Sprintf("references of {%s} where ResultClass = Msvm_SettingsDefineCapabilities", path))
	if err != nil {
		return "", err
	}
	defer enum.Close()

	for {
		entry, err := enum.Next()
		if err != nil {
			return "", err
		}
		if entry == nil {
			return "", errors.New("Could not find settings definition for resource")
		}

		value, vErr := wmiext.GetPropertyAsUint(entry, "ValueRole")
		ref, pErr := entry.GetPropertyAsString("PartComponent")
		entry.Close()
		if vErr == nil && pErr == nil && value == 0 {
			return ref, nil
		}
	}
}
