package e2e

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/containers/libhvee/pkg/hypervctl"
	"github.com/containers/storage/pkg/stringid"
	. "github.com/onsi/ginkgo/v2"
)

const defaultDiskSize = 5

func setDiskDir() string {
	hd, err := os.UserHomeDir()
	if err != nil {
		Fail("unable to determine homedir")
	}
	return filepath.Join(hd, "Downloads")
}

var (
	defaultCacheDirPath = os.TempDir()
	defaultDiskPath     = setDiskDir()
)

type testVM struct {
	name   string
	config *hypervctl.HardwareConfig
	vmm    *hypervctl.VirtualMachineManager
	vm     *hypervctl.VirtualMachine
}

func (t *testVM) stopAndRemove() error {
	err := t.vm.StopWithForce()
	if err != nil && !errors.Is(err, hypervctl.ErrMachineNotRunning) {
		fmt.Println(err)
		return err
	}
	return t.vm.Remove(t.config.DiskPath)
}

func (t *testVM) refresh() error {
	refreshed, err := t.vmm.GetMachine(t.name)
	if err != nil {
		return err
	}
	t.vm = refreshed
	return nil
}

var defaultConfig = hypervctl.HardwareConfig{
	CPUs:     2,
	DiskSize: defaultDiskSize,
	Memory:   4096,
	Network:  false,
}

func (t *testVM) copyCacheDiskToVm() error {
	if len(t.name) < 1 {
		return errors.New("testVM has no name")
	}
	shortID := stringid.TruncateID(t.name)
	imageBaseName := filepath.Base(cachedImagePath)
	destFQPathName := filepath.Join(defaultDiskPath, fmt.Sprintf("%s-%s", shortID, imageBaseName))
	t.config.DiskPath = destFQPathName
	dst, err := os.Create(destFQPathName)
	if err != nil {
		return err
	}
	defer func() {
		_ = dst.Close()
	}()
	return copyWithProgress(cachedImagePath, dst)
}

func newDefaultVM() (*testVM, error) {
	randomName := stringid.GenerateRandomID()
	t := new(testVM)
	t.name = randomName
	newConfig := defaultConfig
	t.config = &newConfig
	if err := t.copyCacheDiskToVm(); err != nil {
		return nil, err
	}
	vmm, vm, err := newVM(randomName, &newConfig)
	t.vmm = vmm
	t.vm = vm
	return t, err
}

func newVM(name string, config *hypervctl.HardwareConfig) (*hypervctl.VirtualMachineManager, *hypervctl.VirtualMachine, error) {
	vmm := hypervctl.NewVirtualMachineManager()
	if err := vmm.NewVirtualMachine(name, config); err != nil {
		return nil, nil, err
	}
	vm, err := vmm.GetMachine(name)
	return vmm, vm, err
}
