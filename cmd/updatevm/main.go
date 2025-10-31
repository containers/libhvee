//go:build windows
// +build windows

package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/containers/libhvee/pkg/hypervctl"
	"github.com/containers/libhvee/pkg/powershell"
)

func main() {

	if err := powershell.HypervAvailable(); err != nil {
		panic(err)
	}

	if !powershell.IsHypervAdministrator() {
		panic(powershell.ErrNotAdministrator)
	}

	if len(os.Args) < 4 {
		fmt.Printf("Usage: %s <vm name> <cores> <static mem>\n\n", os.Args[0])

		return
	}

	vmName := os.Args[1]
	cores, err := strconv.ParseUint(os.Args[2], 0, 64)
	if err != nil {
		panic(err)
	}
	mem, err := strconv.ParseUint(os.Args[3], 0, 64)
	if err != nil {
		panic(err)
	}

	vmm := hypervctl.VirtualMachineManager{}

	vm, err := vmm.GetMachine(vmName)
	if err != nil {
		panic(err)
	}

	err = vm.UpdateProcessorMemSettings(func(ps *hypervctl.ProcessorSettings) {
		ps.Count = int64(cores)
	}, func(ms *hypervctl.MemorySettings) {
		ms.DynamicMemoryEnabled = false
		ms.StartupBytes = mem
		ms.MaximumBytes = mem
		ms.MinimumBytes = mem
	})
	if err != nil {
		panic(err)
	}

}
