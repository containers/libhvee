//go:build windows
// +build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"unicode"

	"github.com/n1hility/hypervctl/pkg/hypervctl"
	"golang.org/x/sys/windows"
)

func main() {
	var err error

	vmms := hypervctl.VirtualMachineManager{}

	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s <vm name> (get|add|rm|edit|put|clear) [<key>] [<value>]\n\n", os.Args[0])
		fmt.Printf("\tget   = get all keys or a specific key\n")
		fmt.Printf("\tadd   = create a key if it doesn't exist\n")
		fmt.Printf("\tedit  = change a key that exists\n")
		fmt.Printf("\tput   = create or edit a key\n")
		fmt.Printf("\trm    = delete key\n")
		fmt.Printf("\tclear = delete everything\n\n")

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
	case "get":
		err = getOperation(vm)
	case "clear":
		err = clearOperation(vm)

	default:
		fmt.Printf("Operation must be get, add, rm, edit, clear, or put\n")
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("KVP failed: %s\n", err.Error())
		os.Exit(1)
	}
}

func getOperation(vm *hypervctl.VirtualMachine) error {
	kvp, err := vm.GetKeyValuePairs()
	if err != nil {
		return err
	}
	if len(os.Args) > 3 {
		key := os.Args[3]
		fmt.Printf("%s = %s\n", key, kvp[key])
		return nil
	}

	for key, value := range kvp {
		fmt.Printf("%s = %v\n", key, value)
	}
	return nil
}

func clearOperation(vm *hypervctl.VirtualMachine) error {
	const lineInputModeFlag uint32 = 0x2
	fmt.Printf("This will delete ALL keys. Are you sure? [y/n] ")
	handle := windows.Handle(os.Stdin.Fd())

	var mode uint32
	err := windows.GetConsoleMode(handle, &mode)
	if err == nil {
		// disable line input for single char reads
		_ = windows.SetConsoleMode(handle, mode & ^lineInputModeFlag)
		defer windows.SetConsoleMode(handle, mode)

	}

loop:
	for {
		b := make([]byte, 1)
		n, err := os.Stdin.Read(b)
		if err != nil {
			return err
		}
		if n < 1 {
			continue
		}
		switch unicode.ToLower(rune(b[0])) {
		case 'y':
			fmt.Printf("y\n")
			break loop
		case 'n':
			fmt.Printf("n\n")
			return errors.New("Aborted by request")
		}
	}

	pairs, err := vm.GetKeyValuePairs()
	if err != nil {
		return err
	}

	count := 0
	for key := range pairs {
		err := vm.RemoveKeyValuePair(key)
		if err != nil {
			fmt.Printf("WARN: could not remove key %q\n", key)
		} else {
			fmt.Print(".")
			count++
			if count%40 == 0 {
				fmt.Println()
			}
		}
	}

	fmt.Printf("\n%d keys deleted!\n", count)
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
