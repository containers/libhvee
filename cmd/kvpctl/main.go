//go:build windows

package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"unicode"

	"github.com/containers/libhvee/pkg/hypervctl"
	"github.com/containers/libhvee/pkg/kvp/ginsu"
	"golang.org/x/sys/windows"
)

type kvpcmd string

const (
	add     kvpcmd = "add"
	addIgn  kvpcmd = "add-ign"
	clear   kvpcmd = "clear"
	edit    kvpcmd = "edit"
	get     kvpcmd = "get"
	put     kvpcmd = "put"
	rm      kvpcmd = "rm"
	unknown kvpcmd = ""
)

func getSubCommand(cmd string) kvpcmd {
	switch cmd {
	case string(add):
		return add
	case string(addIgn):
		return addIgn
	case string(edit):
		return edit
	case string(get):
		return get
	case string(put):
		return put
	case string(rm):
		return rm
	case string(clear):
		return clear
	}
	return unknown
}

func printHelp() {
	fmt.Printf("Usage: %s <vm name> (get|add|add-ign|rm|edit|put|clear) [<key>] [<value>]\n\n", os.Args[0])
	fmt.Printf("\tget   = get all keys or a specific key\n")
	fmt.Printf("\tadd   = create a key if it doesn't exist\n")
	fmt.Printf("\tadd-ign   = split and add key-value pairs for an Ignition config\n")
	fmt.Printf("\tedit  = change a key that exists\n")
	fmt.Printf("\tput   = create or edit a key\n")
	fmt.Printf("\trm    = delete one or more keys\n")
	fmt.Printf("\tclear = delete everything\n\n")
	os.Exit(1)
}

func main() {
	lenArgs := len(os.Args[1:])
	if lenArgs < 2 {
		fmt.Printf("error: virtual machine name or command not provided\n")
		printHelp()
	}

	subCmd := getSubCommand(os.Args[2])
	if subCmd == unknown {
		fmt.Printf("error: unknown command %s\n", os.Args[2])
		printHelp()
	}
	vmms := hypervctl.VirtualMachineManager{}
	if len(os.Args) < 3 {

		return
	}

	vm, err := vmms.GetMachine(os.Args[1])
	if err != nil {
		fmt.Printf("Find machine failed: %s\n", err.Error())
		os.Exit(1)
	}
	verifyArgs(kvpcmd(os.Args[2]), len(os.Args[1:]))

	switch subCmd {
	case add:
		err = vm.AddKeyValuePair(os.Args[3], os.Args[4])
	case rm:
		for _, key := range os.Args[3:] {
			if err := vm.RemoveKeyValuePair(key); err != nil {
				exitOnError(err)
			}
		}
	case edit:
		err = vm.ModifyKeyValuePair(os.Args[3], os.Args[4])
	case put:
		err = vm.PutKeyValuePair(os.Args[3], os.Args[4])
	case get:
		err = getOperation(vm)
	case clear:
		err = clearOperation(vm)
	case addIgn:
		err = addIgnFile(vm, os.Args[3])

	default:
		fmt.Printf("Operation must be get, add, add-ign, rm, edit, clear, or put\n")
		os.Exit(1)
	}
	exitOnError(err)
}

func exitOnError(err error) {
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
			return errors.New("aborted by request")
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

func verifyArgs(op kvpcmd, lenArgs int) {
	switch op {
	case add, put, edit:
		if lenArgs < 4 || lenArgs > 4 {
			printHelp()
		}
	case addIgn:
		if lenArgs < 3 || lenArgs > 3 {
			printHelp()
		}
	case rm:
		// nothing
	case clear:
		if lenArgs > 2 {
			printHelp()
		}
	case get:
		if lenArgs < 2 || lenArgs > 3 {
			printHelp()
		}
	}
}

func addIgnFile(vm *hypervctl.VirtualMachine, inputFilename string) error {
	b, err := os.ReadFile(inputFilename)
	if err != nil {
		return err
	}
	parts, err := ginsu.Dice(bytes.NewReader(b))
	if err != nil {
		return err
	}
	for i, v := range parts {
		key := fmt.Sprintf("ignition.config.%d", i)
		if err := vm.AddKeyValuePair(key, v); err != nil {
			return err
		}
		fmt.Println("added key: ", key)
	}
	return nil
}
