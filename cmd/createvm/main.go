package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/n1hility/hypervctl/pkg/hypervctl"
)

func main() {

	if len(os.Args) < 4 {
		fmt.Printf("Usage: %s <vm name> <vhdx file> <iso file>\n\n", os.Args[0])
		fmt.Printf("Example: \n\t%s my-vm c:\\Users\\Bob\\drive.vhdx c:\\Users\\Bob\\cd.iso\n\n", os.Args[0])

		return
	}

	vmName := os.Args[1]
	vhdxFile := abs(os.Args[2])
	isoFile := abs(os.Args[3])

	vmm := hypervctl.VirtualMachineManager{}

	if fileExists(isoFile) {
		panic(fmt.Errorf("Iso file %q does not exist", isoFile))
	}

	if fileExists(vhdxFile) {
		fmt.Println("VHDX file did not exist, creating:" + vhdxFile)
		if err := vmm.CreateVhdxFile(vhdxFile, 15000*1024*1024); err != nil {
			panic(err)
		}
	}

	// System
	systemSettings, err := hypervctl.NewSystemSettingsBuilder().
		PrepareSystemSettings(vmName, nil).
		PrepareMemorySettings(func(ms *hypervctl.MemorySettings) {
			ms.DynamicMemoryEnabled = true
			ms.VirtualQuantity = 8192 // Startup memory
			ms.Reservation = 1024     // min
			ms.Limit = 16384          // max
		}).
		PrepareProcessorSettings(func(ps *hypervctl.ProcessorSettings) {
			ps.VirtualQuantity = 4 // 4 cores
		}).
		Build()

	if err != nil {
		panic(err)
	}

	// Disks

	err = hypervctl.NewDriveSettingsBuilder(systemSettings).
		AddScsiController().
		AddSyntheticDiskDrive(0).
		DefineVirtualHardDisk(vhdxFile, func(vhdss *hypervctl.VirtualHardDiskStorageSettings) {
			// set extra params like
			// vhdss.IOPSLimit = 5000
		}).
		Finish(). // disk
		Finish(). // drive
		AddSyntheticDvdDrive(1).
		DefineVirtualDvdDisk(isoFile).
		Finish(). // disk
		Finish(). // drive
		Finish(). // controller
		Complete()

	if err != nil {
		panic(err)
	}

	// Network

	err = hypervctl.NewNetworkSettingsBuilder(systemSettings).
		AddSyntheticEthernetPort(nil).
		AddEthernetPortAllocation(""). // "" = connect to default switch
		Finish().                      // allocation
		Finish().                      // port
		Complete()

	if err != nil {
		panic(err)
	}
	vm, err := systemSettings.GetVM()
	if err != nil {
		panic(err)
	}

	if err = vm.AddKeyValuePair("fun", "pair"); err != nil {
		panic(err)
	}

	fmt.Println(vm.Path())

	fmt.Println("Done!")
}

func fileExists(file string) bool {
	_, err := os.Stat(file)
	return os.IsNotExist(err)
}

func abs(file string) string {
	path, err := filepath.Abs(file)
	if err != nil {
		panic(fmt.Errorf("Error building path for %s: %w", path, err))
	}

	return path
}
