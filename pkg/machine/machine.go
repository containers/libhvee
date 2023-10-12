//go:build windows
// +build windows

//nolint:unused
package hypervctl

import (
	"errors"
	"fmt"
	"strings"

	"github.com/containers/libhvee/pkg/hypervctl"
)

// Machine represents the information needed to be a
// podman machine
type Machine struct {
	diskSize     uint64
	ignitionPath string
	imagePath    string
	memory       int64
	name         string
	processors   uint
	timezone     string
	username     string
	vm           *hypervctl.VirtualMachine
	volumes      []hyperVVolume
}

// hyperVVolume is a typed string representation of a mount
type hyperVVolume struct {
	combinedPath string
	guestPath    string
	hostPath     string
}

var ErrNotImplemented = errors.New("function not implemented")

func newHyperVVolume(path string) (*hyperVVolume, error) {
	vol := &hyperVVolume{combinedPath: path}
	return vol.validate()
}

func (v *hyperVVolume) validate() (*hyperVVolume, error) {
	// todo, check each side of the split starts with / ?
	// todo, should we check the host side to make sure it exists?
	split := strings.Split(v.combinedPath, ":")
	if len(split) != 2 {
		return nil, fmt.Errorf("volume path %s is invalid", v.combinedPath)
	}
	return v, nil
}

func NewMachine(name string) (*Machine, error) {
	// TODO Check to make sure name is not used, return error
	h := Machine{name: name}
	return &h, nil
}

func (m *Machine) withDiskSize(size uint64) *Machine {
	m.diskSize = size
	return m
}

func (m *Machine) withIgnitionPath(path string) *Machine {
	m.ignitionPath = path
	return m
}

func (m *Machine) withImagePath(path string) *Machine {
	m.imagePath = path
	return m
}

func (m *Machine) withMemory(size int64) *Machine {
	m.memory = size
	return m
}

func (m *Machine) withProcessors(processorCount uint) *Machine {
	m.processors = processorCount
	return m
}

func (m *Machine) withTimeZone(tz string) *Machine {
	m.timezone = tz
	return m
}

func (m *Machine) withUsername(name string) *Machine {
	m.username = name
	return m
}

func (m *Machine) withVolume(vol hyperVVolume) *Machine {
	m.volumes = append(m.volumes, vol)
	return m
}

// Create uses the builder to actually create the new virtual machine in hyperv
func (m *Machine) Create() (*hypervctl.VirtualMachine, error) {
	return nil, ErrNotImplemented
}

// ChangeDiskSize alters the size of machine's disk
func (m *Machine) ChangeDiskSize(newSize uint64) error {
	return ErrNotImplemented
}

// ChangeMemorySize changes the amount of memory allocated to the machine
func (m *Machine) ChangeMemorySize(newsize uint64) error {
	return ErrNotImplemented
}

// ChangeProcesorCounteranges the number of CPUs assigned to the machine
func (m *Machine) ChangeProcesorCounter(processorCount uint) error {
	return ErrNotImplemented
}
