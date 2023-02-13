//go:build windows
// +build windows

package main

import (
	"bytes"
	"fmt"
	"github.com/containers/libhivee/pkg/ignition"
	"os"

	"github.com/containers/libhivee/pkg/hypervctl"
)

func main() {
	vmms := hypervctl.VirtualMachineManager{}

	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s <vm name> (add|rm|edit|put) <key> [<value>]\n\n", os.Args[0])
		fmt.Printf("\tadd  = create a key if it doesn't exist\n")
		fmt.Printf("\tedit = change a key that exists\n")
		fmt.Printf("\tput  = create or edit a key\n")
		fmt.Printf("\trm   = delete key\n\n")
		return
	}

	vm, err := vmms.GetMachine(os.Args[1])
	if err != nil {
		fmt.Printf("Find machine failed: %s\n", err.Error())
		os.Exit(1)
	}

	switch os.Args[2] {
	case "add":
		verifyArgs("add", true)
		err = vm.AddKeyValuePair(os.Args[3], os.Args[4])
	case "rm":
		verifyArgs("rm", false)
		err = vm.RemoveKeyValuePair(os.Args[3])
	case "edit":
		verifyArgs("edit", false)
		err = vm.ModifyKeyValuePair(os.Args[3], os.Args[4])
	case "put":
		verifyArgs("put", false)
		err = vm.PutKeyValuePair(os.Args[3], os.Args[4])
	case "add-ign":
		err = addIgnFile(vm, os.Args[3])
	default:
		fmt.Printf("Operation must be add, rm, edit, or put\n")
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("KVP failed: %s\n", err.Error())
		os.Exit(1)
	}
}

func addIgnFile(vm *hypervctl.VirtualMachine, name string) error {
	b, err := os.ReadFile(name)
	if err != nil {
		return err
	}
	parts, err := ignition.Dice(bytes.NewReader(b))
	if err != nil {
		return err
	}
	for i, v := range parts {
		key := fmt.Sprintf("%s%d", ignition.Key, i)
		if err := vm.AddKeyValuePair(key, string(v)); err != nil {
			return err
		}
		fmt.Println("added key: ", key)
	}
	return nil
}

func verifyArgs(operation string, value bool) {
	check := 4
	suffix := ""

	if value {
		check++
		suffix = " <value>"
	}

	if len(os.Args) < check {
		fmt.Printf("Usage: %s <vm name> %s <key>%s", os.Args[0], operation, suffix)
		os.Exit(1)
	}
}
