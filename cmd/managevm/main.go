//go:build windows

package main

import (
	"fmt"
	"os"

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

	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <vm name> <start|stop|restart|status|remove> \n\n", os.Args[0])
		return
	}

	vmName := os.Args[1]
	action := os.Args[2]

	vmm := hypervctl.VirtualMachineManager{}

	exists, _, err := vmm.GetMachineExists(vmName)
	if err != nil {
		panic(err)
	}
	if !exists {
		panic(fmt.Errorf("VM %s does not exist", vmName))
	}

	vm, err := vmm.GetMachine(vmName)
	if err != nil {
		panic(err)
	}

	switch action {
	case "remove":
		err = vm.Remove(vm.HardDrives[0].Path)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
	case "start":
		err = vm.Start()
	case "stop":
		err = vm.Stop()
	case "restart":
		err = vm.Stop()
		if err != nil {
			panic(err)
		}
		err = vm.Start()
		if err != nil {
			panic(err)
		}
	case "status":
		fmt.Println(vm.GetState())
	}
	if err != nil {
		panic(err)
	}
}
