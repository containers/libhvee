//go:build windows
// +build windows

package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/containers/libhvee/pkg/hypervctl"
)

func main() {

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
		ps.VirtualQuantity = cores
	}, func(ms *hypervctl.MemorySettings) {
		ms.DynamicMemoryEnabled = false
		ms.VirtualQuantity = mem
		ms.Limit = mem
		ms.Reservation = mem
	})
	if err != nil {
		panic(err)
	}

}
