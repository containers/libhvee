//go:build windows
// +build windows

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/containers/libhvee/pkg/hypervctl"
)

const (
	summary = "summary"
	vms     = "vms"
)

func getVms() (string, error) {
	vmms := hypervctl.VirtualMachineManager{}
	vms, err := vmms.GetAll()
	if err != nil {
		return "", fmt.Errorf("Could not retrieve virtual machines: %s\n", err.Error())
	}
	b, err := json.MarshalIndent(vms, "", "\t")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func dumpSummary() (string, error) {
	vmms := hypervctl.VirtualMachineManager{}
	summs, err := vmms.GetSummaryInformation(hypervctl.SummaryRequestNearAll)
	if err != nil {
		return "", fmt.Errorf("Could not retrieve virtual machine summaries: %v\n", err)
	}
	b, err := json.MarshalIndent(summs, "", "\t")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func printHelp() {
	fmt.Printf("argument must be one of %q or %q", summary, vms)
}

func main() {
	var (
		err    error
		result string
	)
	args := os.Args
	if len(args) != 2 {
		printHelp()
		os.Exit(1)
	}
	if arg := args[1]; arg != summary && arg != vms {
		printHelp()
		os.Exit(1)
	}
	if args[1] == summary {
		result, err = dumpSummary()
	} else {
		result, err = getVms()
	}
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(result)
}
